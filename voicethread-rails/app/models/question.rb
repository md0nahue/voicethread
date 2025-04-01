class Question < ApplicationRecord
  belongs_to :interview_section
  belongs_to :topic

  after_create :enqueue_audio_generation

  def generate_audio_url
    update(status: 'processing')
    
    polly = Aws::Polly::Client.new
    response = polly.synthesize_speech({
      output_format: "mp3",
      text: body,
      voice_id: "Joanna"
    })

    # Generate a unique filename
    filename = "questions/#{id}_#{Time.now.to_i}.mp3"
    
    # Save the audio stream to a file
    File.open(filename, "wb") do |file|
      file.write(response.audio_stream.read)
    end

    # Calculate audio duration using ffmpeg
    duration = calculate_audio_duration(filename)

    # Upload to S3 and get URL
    s3 = Aws::S3::Resource.new
    bucket = s3.bucket(ENV['AWS_BUCKET'])
    obj = bucket.object(filename)
    obj.upload_file(filename)

    # Save the URL, duration, and update status
    update(
      url: obj.public_url,
      duration: duration,
      status: 'completed'
    )

    # Clean up local file
    File.delete(filename)
  rescue StandardError => e
    update(status: 'failed')
    Rails.logger.error("Failed to generate audio for question #{id}: #{e.message}")
  end

  private

  def enqueue_audio_generation
    GenerateAudioJob.perform_later(id)
  end

  def calculate_audio_duration(filename)
    # Use ffmpeg to get audio duration in milliseconds
    cmd = "ffmpeg -i #{filename} 2>&1 | grep 'Duration' | cut -d ' ' -f 4 | sed s/,//"
    duration_str = `#{cmd}`.strip
    
    # Convert HH:MM:SS.mmm to milliseconds
    hours, minutes, seconds = duration_str.split(':').map(&:to_f)
    (hours * 3600 + minutes * 60 + seconds) * 1000
  end
end 