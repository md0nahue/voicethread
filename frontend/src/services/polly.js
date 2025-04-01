import { Polly } from '@aws-sdk/client-polly';

const polly = new Polly({
  region: import.meta.env.VITE_AWS_REGION,
  credentials: {
    accessKeyId: import.meta.env.VITE_AWS_ACCESS_KEY_ID,
    secretAccessKey: import.meta.env.VITE_AWS_SECRET_ACCESS_KEY,
  },
});

export async function synthesizeSpeech(text) {
  try {
    const params = {
      Engine: 'neural',
      LanguageCode: 'en-US',
      Text: text,
      OutputFormat: 'mp3',
      VoiceId: 'Matthew',
    };

    const command = new SynthesizeSpeechCommand(params);
    const response = await polly.send(command);

    // Convert the audio stream to a blob
    const blob = new Blob([await response.AudioStream.transformToByteArray()], {
      type: 'audio/mpeg',
    });

    return URL.createObjectURL(blob);
  } catch (error) {
    console.error('Error synthesizing speech:', error);
    throw error;
  }
} 