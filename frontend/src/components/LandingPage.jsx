import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

export default function LandingPage() {
  const [topic, setTopic] = useState('');
  const navigate = useNavigate();

  const handleStartInterview = () => {
    if (topic.trim()) {
      const sessionId = crypto.randomUUID();
      navigate(`/recorder?session=${sessionId}&topic=${encodeURIComponent(topic)}`);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-lg shadow-lg">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">VoiceThread</h1>
          <p className="text-gray-600">What do you want to be interviewed about?</p>
        </div>
        
        <div className="mt-8 space-y-6">
          <div>
            <input
              type="text"
              value={topic}
              onChange={(e) => setTopic(e.target.value)}
              placeholder="Enter your topic..."
              className="w-full px-4 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary focus:border-primary"
            />
          </div>
          
          <button
            onClick={handleStartInterview}
            disabled={!topic.trim()}
            className="w-full py-3 px-4 bg-primary text-white rounded-md hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Start Interview
          </button>
        </div>
      </div>
    </div>
  );
} 