import mongoose from 'mongoose';

const audioSchema = new mongoose.Schema({
  sessionId: {
    type: String,
    required: true,
    index: true
  },
  topic: {
    type: String,
    required: true
  },
  chunks: [{
    timestamp: Date,
    audioUrl: String,
    transcription: String,
    polishedTranscription: String
  }],
  status: {
    type: String,
    enum: ['recording', 'processing', 'completed'],
    default: 'recording'
  },
  createdAt: {
    type: Date,
    default: Date.now
  }
});

export default mongoose.model('Audio', audioSchema); 