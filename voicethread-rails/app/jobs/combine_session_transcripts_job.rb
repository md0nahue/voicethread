class CombineSessionTranscriptsJob < ApplicationJob
  queue_as :default

  def perform(session_id)
    session = InterviewSession.find(session_id)
    return if session.raw_transcript.present?

    # Get all completed audio chunks with transcriptions
    chunks = AudioChunk.where(session_id: session_id)
                      .where(status: 'complete')
                      .where.not(transcription: nil)
                      .order(:chunk_number)
                      .to_a

    return if chunks.empty?

    # Combine all transcriptions with timestamps
    raw_transcript = chunks.map do |chunk|
      timestamp = format_timestamp(chunk.created_at)
      "[#{timestamp}] #{chunk.transcription}"
    end.join("\n")

    # Update the session with the raw transcript
    session.update!(raw_transcript: raw_transcript)
  end

  private

  def format_timestamp(time)
    time.strftime("%H:%M:%S")
  end
end 