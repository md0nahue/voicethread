class Topic < ApplicationRecord
  after_create :generate_questions
  has_many :interview_sections
  has_many :questions, through: :interview_sections

  def add_fee(fee)
    # Initialize fees array if nil
    self.fees ||= []
    # Add new fee to array
    self.fees << fee
    save
  end

  def total_fee
    self.fees.sum
  end

  def generate_questions
    instructions = <<-INSTRUCTIONS
    You are an AI assistant that responds in a structured JSON format. Always return an array of strings in your responses. Example:
    ["string1", "string2", "string3"]
    INSTRUCTIONS

    prompt = <<-PROMPT
      Brainstorm an list of exactly 15 high-level topics for an interview
      about the topic listed below:
      #{body}
    PROMPT

    response = AskAi.ask(instructions, prompt)
    answer = response[:answer]
    fee = response[:fee]
    add_fee(fee)
    
    # Create sections and generate their questions
    answer.each do |_body|
      section = InterviewSection.create!(body: _body, topic: self)
      section.generate_questions
    end
  end
end 