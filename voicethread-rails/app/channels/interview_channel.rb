class InterviewChannel < ApplicationCable::Channel
  def subscribed
    stream_for current_user
  end

  def unsubscribed
    stop_all_streams
  end

  def request_questions(data)
    topic = current_user.topics.find(data['topic_id'])
    
    # Format the questions data
    questions_data = {
      topic_id: topic.id,
      topic_body: topic.body,
      sections: topic.interview_sections.map do |section|
        {
          id: section.id,
          body: section.body,
          questions: section.questions.map do |question|
            {
              id: question.id,
              body: question.body,
              audio_url: question.url
            }
          end
        }
      end
    }

    # Broadcast to the Golang server's WebSocket endpoint
    InterviewChannel.broadcast_to(
      current_user,
      {
        type: 'interview_questions',
        data: questions_data
      }
    )
  end
end 