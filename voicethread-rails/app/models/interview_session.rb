class InterviewSession < ApplicationRecord
  belongs_to :user
  belongs_to :topic
  has_many :audio_chunks, dependent: :destroy

  validates :status, inclusion: { in: %w[in_progress completed] }
  validates :final_audio_url, presence: true, if: :completed?

  scope :completed, -> { where(status: 'completed') }
  scope :in_progress, -> { where(status: 'in_progress') }

  def completed?
    status == 'completed'
  end

  def final_audio_url
    return nil unless super
    "https://#{ENV['AWS_BUCKET_NAME']}.s3.amazonaws.com/#{super}"
  end

  def enqueue_audio_combination
    CombineSessionAudioJob.perform_later(id)
  end

  def combine_audio_chunks
    enqueue_audio_combination
  end

  def enqueue_transcription_polishing
    PolishSessionTranscriptionJob.perform_later(id)
  end

  def polish_transcription
    enqueue_transcription_polishing
  end

  def enqueue_transcript_combination
    CombineSessionTranscriptsJob.perform_later(id)
  end

  def combine_transcripts
    enqueue_transcript_combination
  end

  def process_completion
    combine_audio_chunks
    combine_transcripts
    polish_transcription
  end
end 