class PolishSessionTranscriptionJob < ApplicationJob
  queue_as :default

  def perform(session_id)
    session = InterviewSession.find(session_id)
    return if session.polished_transcription.present?

    # Get all completed audio chunks with transcriptions
    chunks = AudioChunk.where(session_id: session_id)
                      .where(status: 'complete')
                      .where.not(transcription: nil)
                      .order(:chunk_number)
                      .to_a

    return if chunks.empty?

    # Combine all transcriptions with timestamps
    raw_transcription = chunks.map do |chunk|
      timestamp = format_timestamp(chunk.created_at)
      "##{timestamp}\n#{chunk.transcription}"
    end.join("\n\n")

    # Polish the transcription using RubyLLM
    polished_text = polish_transcription(raw_transcription)

    # Update the session with the polished transcription
    session.update!(polished_transcription: polished_text)
  end

  private

  def format_timestamp(time)
    time.strftime("%H:%M:%S")
  end

  def polish_transcription(raw_text)
    client = RubyLLM::Client.new(
      api_key: ENV['OPENAI_API_KEY'],
      model: 'gpt-4-turbo-preview'
    )

    prompt = <<~PROMPT
      Please polish and format the following interview transcription. The text is already in chronological order with timestamps.
      Make it more readable while preserving the original meaning and structure. Format it in Markdown.
      Keep the timestamps but make them more subtle.
      Add appropriate headings and sections if needed.
      Fix any obvious transcription errors or formatting issues.
      Do not add any content that wasn't in the original transcription.

      Transcription:
      #{raw_text}
    PROMPT

    response = client.chat(
      messages: [
        {
          role: 'system',
          content: 'You are a professional transcription editor. Your task is to polish and format interview transcriptions while maintaining their original content and meaning.'
        },
        {
          role: 'user',
          content: prompt
        }
      ],
      temperature: 0.3
    )

    response.choices.first.message.content
  end
end 