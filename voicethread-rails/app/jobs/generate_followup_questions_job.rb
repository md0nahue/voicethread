class GenerateFollowupQuestionsJob < ApplicationJob
  queue_as :default

  def perform(session_id)
    # Get all transcribed chunks that haven't generated followups yet
    chunks = AudioChunk.where(session_id: session_id)
                      .where.not(transcription: nil)
                      .where(created_followup_questions: false)
                      .order(:chunk_number)

    return if chunks.empty?

    # Combine transcriptions
    combined_transcription = chunks.map(&:transcription).join("\n\n")

    # Generate followup questions using RubyLLM
    client = RubyLLM::Client.new(
      api_key: ENV['OPENAI_API_KEY'],
      model: 'gpt-4-turbo-preview'
    )

    response = client.chat(
      messages: [
        {
          role: 'system',
          content: 'You are a helpful AI assistant that creates followup questions based on interview content. Return only an array of strings, with each string being a followup question. Do not include any other text or formatting.'
        },
        {
          role: 'user',
          content: "Create additional followup questions based on this interview content:\n\n#{combined_transcription}"
        }
      ],
      temperature: 0.7,
      response_format: { type: 'json_object' }
    )

    # Parse the response to get the questions array
    questions = JSON.parse(response.choices.first.message.content)['questions']

    # Get the topic from the first chunk's metadata
    topic_id = chunks.first.metadata['topic_id']
    topic = Topic.find(topic_id)

    # Create new questions
    questions.each do |question_text|
      topic.questions.create!(
        body: question_text,
        is_followup_question: true
      )
    end

    # Mark chunks as processed
    chunks.update_all(created_followup_questions: true)
  rescue StandardError => e
    Rails.logger.error("Failed to generate followup questions for session #{session_id}: #{e.message}")
    Rails.logger.error(e.backtrace.join("\n"))
  end
end 