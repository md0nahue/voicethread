class InterviewSection < ApplicationRecord
  belongs_to :topic
  has_many :questions

  def generate_questions
    instructions = <<-INSTRUCTIONS
    You are an AI assistant that responds in a structured JSON format. Always return an array of strings in your responses. Example:
    ["string1", "string2", "string3"]
    INSTRUCTIONS

    prompt = <<-PROMPT
      Brainstorm an list of exactly 15 unique interview questions about:
      #{body} for #{topic.body}
    PROMPT

    response = AskAi.ask(instructions, prompt)
    answer = response[:answer]
    fee = response[:fee]
    topic.add_fee(fee)

    answer.each do |_body|
      Question.create!(body: _body, interview_section: self)
    end
  end
end 