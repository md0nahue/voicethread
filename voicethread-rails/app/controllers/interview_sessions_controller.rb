class InterviewSessionsController < ApplicationController
  before_action :authenticate_user!
  before_action :set_interview_session, only: [:show]

  def index
    @interview_sessions = current_user.interview_sessions
                                    .includes(:topic)
                                    .order(created_at: :desc)
  end

  def show
    authorize @interview_session
  end

  private

  def set_interview_session
    @interview_session = InterviewSession.find(params[:id])
  end
end 