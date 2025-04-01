class CreateAudioChunks < ActiveRecord::Migration[7.1]
  def change
    create_table :audio_chunks do |t|
      t.string :session_id, null: false
      t.string :s3_key, null: false
      t.integer :chunk_number, null: false
      t.integer :duration, null: false  # Duration in milliseconds
      t.integer :size, null: false      # File size in bytes
      t.string :status, default: 'new', null: false
      t.text :transcription
      t.jsonb :metadata, default: {}
      t.timestamps

      t.index [:session_id, :chunk_number], unique: true
      t.index :status
    end
  end
end 