module Api
  module V1
    class TopicsController < ApplicationController
      before_action :authenticate_user!
      before_action :set_topic, only: [:show]

      def index
        @topics = current_user.topics
        render json: @topics.map { |topic| format_topic(topic) }
      end

      def show
        render json: format_topic(@topic)
      end

      private

      def set_topic
        @topic = current_user.topics.find(params[:id])
      end

      def format_topic(topic)
        {
          id: topic.id,
          body: topic.body,
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
    end
  end
end 