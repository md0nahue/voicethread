import React, { useState, useEffect, useRef } from 'react';
import { useSearchParams } from 'react-router-dom';
import WaveSurfer from 'wavesurfer.js';
import RecordPlugin from 'wavesurfer.js/dist/plugins/record.js';

const WS_URL = 'ws://localhost:8080/ws';
const SILENCE_THRESHOLD = 0.01;
const SILENCE_DURATION = 1000; // 1 second of silence to trigger new chunk

const Recorder = ({ sessionId, onSessionComplete }) => {
  const [searchParams] = useSearchParams();
  const [isRecording, setIsRecording] = useState(false);
  const [isPlaying, setIsPlaying] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [sessionStatus, setSessionStatus] = useState('in_progress');
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState(null);
  const [audioChunks, setAudioChunks] = useState([]);
  const [isSilenceDetected, setIsSilenceDetected] = useState(true);
  const [questions, setQuestions] = useState([]);
  const [currentQuestion, setCurrentQuestion] = useState(null);
  const mediaRecorderRef = useRef(null);
  const timerRef = useRef(null);
  const waveformRef = useRef(null);
  const wavesurferRef = useRef(null);
  const recordPluginRef = useRef(null);
  const wsRef = useRef(null);
  const silenceTimerRef = useRef(null);
  const currentChunkRef = useRef([]);
  const topicId = searchParams.get('topic');

  useEffect(() => {
    // Initialize WebSocket connection
    wsRef.current = new WebSocket(WS_URL);

    wsRef.current.onopen = () => {
      console.log('WebSocket connected');
      // Request questions when connected
      if (topicId) {
        requestQuestions();
      }
      // Send initial session info
      wsRef.current.send(JSON.stringify({
        type: 'session_start',
        session_id: sessionId
      }));
    };

    wsRef.current.onmessage = (event) => {
      const data = JSON.parse(event.data);
      handleWebSocketMessage(data);
    };

    wsRef.current.onerror = (error) => {
      console.error('WebSocket error:', error);
      setError('Connection error occurred');
    };

    wsRef.current.onclose = () => {
      console.log('WebSocket disconnected');
    };

    // Initialize WaveSurfer with Record plugin
    wavesurferRef.current = WaveSurfer.create({
      container: waveformRef.current,
      height: 100,
      waveColor: '#4F46E5',
      progressColor: '#10B981',
      barWidth: 2,
      barRadius: 3,
      cursorWidth: 1,
      cursorColor: '#4F46E5',
      barGap: 3,
      normalize: true,
      fillParent: true
    });

    // Initialize Record plugin
    recordPluginRef.current = wavesurferRef.current.registerPlugin(
      RecordPlugin.create({
        scrollingWaveform: true,
        renderRecordedAudio: false,
        audioBitsPerSecond: 128000
      })
    );

    // Set up audio processing for silence detection
    recordPluginRef.current.on('record-progress', (duration) => {
      const audioData = recordPluginRef.current.getAnalyser().getFloatTimeDomainData();
      const isSilent = audioData.every(value => Math.abs(value) < SILENCE_THRESHOLD);
      
      if (isSilent) {
        if (!silenceTimerRef.current) {
          silenceTimerRef.current = setTimeout(() => {
            setIsSilenceDetected(true);
            if (currentChunkRef.current.length > 0) {
              // Send current chunk and silence detection message
              sendAudioChunk(new Blob(currentChunkRef.current, { type: 'audio/webm;codecs=opus' }));
              sendSilenceDetected();
              currentChunkRef.current = [];
            }
          }, SILENCE_DURATION);
        }
      } else {
        if (silenceTimerRef.current) {
          clearTimeout(silenceTimerRef.current);
          silenceTimerRef.current = null;
        }
        setIsSilenceDetected(false);
      }
    });

    wavesurferRef.current.on('ready', () => {
      setDuration(wavesurferRef.current.getDuration());
    });

    wavesurferRef.current.on('audioprocess', () => {
      setCurrentTime(wavesurferRef.current.getCurrentTime());
    });

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
      if (wavesurferRef.current) {
        wavesurferRef.current.destroy();
      }
      if (timerRef.current) {
        clearInterval(timerRef.current);
      }
      if (silenceTimerRef.current) {
        clearTimeout(silenceTimerRef.current);
      }
      if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
        mediaRecorderRef.current.stop();
      }
    };
  }, [sessionId, topicId]);

  const requestQuestions = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({
        type: 'request_questions',
        topicId: topicId,
        userId: 'current-user-id' // TODO: Get actual user ID
      }));
    }
  };

  const handleQuestionsReceived = (response) => {
    // Flatten all questions from all sections
    const allQuestions = response.sections.flatMap(section => 
      section.questions.map(q => ({
        ...q,
        sectionId: section.id,
        sectionBody: section.body
      }))
    );
    setQuestions(allQuestions);
    
    // Set the first question as current if not already set
    if (!currentQuestion && allQuestions.length > 0) {
      setCurrentQuestion(allQuestions[0]);
    }
  };

  const handleWebSocketMessage = (data) => {
    switch (data.type) {
      case 'ack':
        console.log('Received acknowledgment:', data);
        break;
      case 'status_update':
        handleStatusUpdate(data);
        break;
      case 'session_complete':
        handleSessionComplete(data);
        break;
      case 'interview_questions':
        handleQuestionsReceived(data);
        break;
      default:
        console.log('Unknown message type:', data.type);
    }
  };

  const handleStatusUpdate = (data) => {
    setSessionStatus(data.status);
    setProgress(data.progress);
    if (data.message) {
      console.log('Status message:', data.message);
    }
  };

  const handleSessionComplete = (data) => {
    setSessionStatus('completed');
    setProgress(100);
    if (onSessionComplete) {
      onSessionComplete(data);
    }
  };

  const sendAudioChunk = async (blob) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const message = {
        type: 'audio',
        sessionId: sessionId,
        data: await blob.arrayBuffer()
      };
      wsRef.current.send(JSON.stringify(message));
    }
  };

  const sendSilenceDetected = () => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      const message = {
        type: 'silence',
        sessionId: sessionId
      };
      wsRef.current.send(JSON.stringify(message));
    }
  };

  const startRecording = async () => {
    try {
      // Start recording using WaveSurfer's Record plugin
      await recordPluginRef.current.startRecording();
      setIsRecording(true);
      currentChunkRef.current = [];

      // Set up the MediaRecorder for collecting chunks
      const stream = recordPluginRef.current.getMediaStream();
      const mediaRecorder = new MediaRecorder(stream, {
        mimeType: 'audio/webm;codecs=opus'
      });
      mediaRecorderRef.current = mediaRecorder;

      mediaRecorder.ondataavailable = (event) => {
        if (event.data.size > 0) {
          currentChunkRef.current.push(event.data);
        }
      };

      mediaRecorder.start(100); // Collect data every 100ms
      timerRef.current = setInterval(() => {
        setCurrentTime((prev) => prev + 1);
      }, 1000);
    } catch (error) {
      console.error('Error starting recording:', error);
      setError('Failed to start recording');
    }
  };

  const stopRecording = async () => {
    if (recordPluginRef.current) {
      recordPluginRef.current.stopRecording();
    }
    if (mediaRecorderRef.current && mediaRecorderRef.current.state === 'recording') {
      mediaRecorderRef.current.stop();
      mediaRecorderRef.current.stream.getTracks().forEach(track => track.stop());
      
      // Send any remaining audio data
      if (currentChunkRef.current.length > 0) {
        await sendAudioChunk(new Blob(currentChunkRef.current, { type: 'audio/webm;codecs=opus' }));
        currentChunkRef.current = [];
      }
    }
    setIsRecording(false);
    if (timerRef.current) {
      clearInterval(timerRef.current);
    }
    if (silenceTimerRef.current) {
      clearTimeout(silenceTimerRef.current);
    }
  };

  const formatTime = (seconds) => {
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = Math.floor(seconds % 60);
    return `${minutes}:${remainingSeconds.toString().padStart(2, '0')}`;
  };

  const togglePlayPause = () => {
    if (wavesurferRef.current) {
      wavesurferRef.current.playPause();
      setIsPlaying(!isPlaying);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-8">
      <div className="max-w-2xl mx-auto bg-white rounded-lg shadow-lg p-6">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-2xl font-bold text-gray-900">🎙️ VoiceThread Recorder</h1>
          <div className="text-sm text-gray-500">
            Session: {sessionId}
          </div>
        </div>

        {currentQuestion && (
          <div className="mb-6">
            <h2 className="text-lg font-medium text-gray-700 mb-2">Current Question:</h2>
            <p className="text-gray-600">{currentQuestion.body}</p>
            {currentQuestion.is_followup && (
              <span className="inline-block mt-2 px-2 py-1 text-xs font-semibold text-blue-800 bg-blue-100 rounded-full">
                Follow-up Question
              </span>
            )}
          </div>
        )}

        <div className="mb-6 border rounded-lg p-4 bg-gray-50">
          <div ref={waveformRef} />
        </div>

        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-4">
            <button
              onClick={togglePlayPause}
              className={`px-4 py-2 rounded-md text-white ${
                isPlaying ? 'bg-red-500 hover:bg-red-600' : 'bg-primary hover:bg-primary/90'
              }`}
            >
              {isPlaying ? 'Pause' : 'Play'}
            </button>
            <span className="text-gray-600">
              {formatTime(currentTime)} / {formatTime(duration)}
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <span className={`px-3 py-1 rounded-full text-sm ${
              sessionStatus === 'completed' ? 'bg-green-100 text-green-800' :
              sessionStatus === 'processing' ? 'bg-yellow-100 text-yellow-800' :
              'bg-blue-100 text-blue-800'
            }`}>
              {sessionStatus.charAt(0).toUpperCase() + sessionStatus.slice(1)}
            </span>
            {progress > 0 && (
              <span className="text-sm text-gray-600">
                {Math.round(progress)}%
              </span>
            )}
          </div>
        </div>

        <div className="flex items-center justify-between text-sm text-gray-500">
          <div>🔌 {wsRef.current?.readyState === WebSocket.OPEN ? 'WebSocket Connected' : 'WebSocket Disconnected'}</div>
          <div>🎤 {isRecording ? 'Mic Active' : 'Mic Inactive'}</div>
          <div>📦 Chunks Sent: {audioChunks.length}</div>
        </div>

        <div className="flex justify-center space-x-4">
          <button
            onClick={isRecording ? stopRecording : startRecording}
            className={`px-4 py-2 rounded-md text-white font-medium ${
              isRecording
                ? 'bg-red-600 hover:bg-red-700'
                : 'bg-blue-600 hover:bg-blue-700'
            }`}
          >
            {isRecording ? 'Stop Recording' : 'Start Recording'}
          </button>
        </div>

        {error && (
          <div className="mt-4 p-4 bg-red-50 text-red-700 rounded-md">
            {error}
          </div>
        )}
      </div>
    </div>
  );
};

export default Recorder; 