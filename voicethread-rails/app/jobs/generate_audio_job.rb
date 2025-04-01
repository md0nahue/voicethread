class GenerateAudioJob < ApplicationJob
  queue_as :default

  def perform(question_id)
    question = Question.find(question_id)
    question.generate_audio_url
  end
end 