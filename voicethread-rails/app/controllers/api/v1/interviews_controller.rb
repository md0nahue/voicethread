module Api
  module V1
    class InterviewsController < ApplicationController
      skip_before_action :verify_authenticity_token
      before_action :authenticate_user_from_token

      def questions
        topic = current_user.topics.find(params[:topic_id])
        
        render json: {
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
      end

      private

      def authenticate_user_from_token
        token = request.headers['Authorization']&.split(' ')&.last
        return head :unauthorized unless token

        user = User.find_by_auth_token(token)
        head :unauthorized unless user
        @current_user = user
      end
    end
  end
end 