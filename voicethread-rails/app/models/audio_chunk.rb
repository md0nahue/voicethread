class AudioChunk < ApplicationRecord
  validates :session_id, presence: true
  validates :s3_key, presence: true
  validates :chunk_number, presence: true, numericality: { only_integer: true, greater_than: 0 }
  validates :duration, presence: true, numericality: { only_integer: true, greater_than: 0 }
  validates :size, presence: true, numericality: { only_integer: true, greater_than: 0 }
  validates :status, presence: true, inclusion: { in: %w[new transcribed] }
  validates :session_id, uniqueness: { scope: :chunk_number }

  # Scopes for common queries
  scope :new_chunks, -> { where(status: 'new') }
  scope :transcribed, -> { where(status: 'transcribed') }
  scope :for_session, ->(session_id) { where(session_id: session_id) }
  scope :ordered, -> { order(chunk_number: :asc) }

  # Get the full S3 URL for this chunk
  def s3_url
    s3 = Aws::S3::Resource.new
    bucket = s3.bucket(ENV['AWS_BUCKET'])
    bucket.object(s3_key).public_url
  end

  # Mark this chunk as transcribed
  def mark_as_transcribed!(transcription_text)
    update!(
      status: 'transcribed',
      transcription: transcription_text
    )
  end
end 