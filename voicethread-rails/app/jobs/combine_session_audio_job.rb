class CombineSessionAudioJob < ApplicationJob
  queue_as :default

  def perform(session_id)
    session = InterviewSession.find(session_id)
    return if session.final_audio_url.present?

    # Get all audio chunks for this session, ordered by chunk number
    chunks = AudioChunk.where(session_id: session_id)
                      .where(status: 'complete')
                      .order(:chunk_number)
                      .to_a

    return if chunks.empty?

    # Create a temporary directory for processing
    temp_dir = Rails.root.join('tmp', 'audio_processing', session_id.to_s)
    FileUtils.mkdir_p(temp_dir)

    begin
      # Download all chunks
      chunks.each do |chunk|
        temp_file = temp_dir.join("chunk_#{chunk.chunk_number}.wav")
        download_from_s3(chunk.s3_key, temp_file)
      end

      # Combine all chunks into one file
      combined_file = temp_dir.join('combined.wav')
      combine_audio_files(chunks.map { |c| temp_dir.join("chunk_#{c.chunk_number}.wav") }, combined_file)

      # Upload the combined file to S3
      final_key = "sessions/#{session_id}/final_audio.wav"
      upload_to_s3(combined_file, final_key)

      # Update the session with the final audio URL and total duration
      total_duration = chunks.sum(&:duration)
      session.update!(
        final_audio_url: final_key,
        total_duration: total_duration,
        status: 'completed'
      )

    ensure
      # Clean up temporary files
      FileUtils.rm_rf(temp_dir) if Dir.exist?(temp_dir)
    end
  end

  private

  def download_from_s3(key, local_path)
    s3 = Aws::S3::Client.new
    response = s3.get_object(bucket: ENV['AWS_BUCKET_NAME'], key: key)
    File.binwrite(local_path, response.body.read)
  end

  def upload_to_s3(local_path, key)
    s3 = Aws::S3::Client.new
    File.open(local_path, 'rb') do |file|
      s3.put_object(
        bucket: ENV['AWS_BUCKET_NAME'],
        key: key,
        body: file,
        content_type: 'audio/wav'
      )
    end
  end

  def combine_audio_files(input_files, output_file)
    # Use ffmpeg to combine the audio files
    input_args = input_files.map { |f| "-i #{f}" }.join(' ')
    filter_complex = (0...input_files.length).map { |i| "[#{i}:0]" }.join('') + 
                    "concat=n=#{input_files.length}:v=0:a=1[out]"
    
    system("ffmpeg #{input_args} -filter_complex '#{filter_complex}' -map '[out]' #{output_file}")
  end
end 