class WebhooksController < ApplicationController
  skip_before_action :verify_authenticity_token

  def golang
    session = InterviewSession.find(params[:session_id])
    session.update!(
      status: params[:status],
      total_duration: params[:duration]
    )

    if params[:status] == 'completed'
      session.process_completion
    end

    head :ok
  end
end
