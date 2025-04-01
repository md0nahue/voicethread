class AskAi
  def self.ask(instructions, prompt)
    client = OpenAI::Client.new

    response = client.chat(
      parameters: {
        model: "gpt-4",
        messages: [
          { role: "system", content: instructions },
          { role: "user", content: prompt }
        ],
        temperature: 0.7,
        max_tokens: 1000
      }
    )

    {
      answer: JSON.parse(response.dig("choices", 0, "message", "content")),
      fee: calculate_fee(response)
    }
  end

  private

  def self.calculate_fee(response)
    # Calculate fee based on token usage
    # This is a simplified calculation - adjust based on your needs
    tokens = response.dig("usage", "total_tokens")
    (tokens * 0.00003).round(7) # Example rate: $0.00003 per token
  end
end 