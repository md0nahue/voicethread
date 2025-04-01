import express from 'express';
import multer from 'multer';
import path from 'path';
import { fileURLToPath } from 'url';
import { dirname } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const router = express.Router();

// Configure multer for audio file storage
const storage = multer.diskStorage({
  destination: function (req, file, cb) {
    cb(null, path.join(__dirname, '../../uploads/audio'));
  },
  filename: function (req, file, cb) {
    const sessionId = req.body.sessionId || 'unknown';
    const timestamp = Date.now();
    cb(null, `${sessionId}-${timestamp}${path.extname(file.originalname)}`);
  }
});

const upload = multer({ 
  storage: storage,
  limits: {
    fileSize: 10 * 1024 * 1024 // 10MB limit
  },
  fileFilter: (req, file, cb) => {
    if (file.mimetype.startsWith('audio/')) {
      cb(null, true);
    } else {
      cb(new Error('Only audio files are allowed'));
    }
  }
});

// Routes
router.post('/upload', upload.single('audio'), async (req, res) => {
  try {
    if (!req.file) {
      return res.status(400).json({ error: 'No audio file provided' });
    }

    const audioUrl = `/audio/${req.file.filename}`;
    
    res.status(200).json({
      success: true,
      audioUrl,
      filename: req.file.filename,
      sessionId: req.body.sessionId
    });
  } catch (error) {
    console.error('Error uploading audio:', error);
    res.status(500).json({ error: 'Error uploading audio file' });
  }
});

router.get('/session/:sessionId', async (req, res) => {
  try {
    const sessionId = req.params.sessionId;
    // For now, just return success - we'll implement file listing later
    res.status(200).json({
      success: true,
      sessionId,
      message: 'Session found'
    });
  } catch (error) {
    console.error('Error getting session:', error);
    res.status(500).json({ error: 'Error retrieving session data' });
  }
});

export default router; 