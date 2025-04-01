class CreateAudioChunks < ActiveRecord::Migration[7.1]
  def change
    create_table :audio_chunks do |t|
      t.string :session_id, null: false
      t.string :s3_key, null: false
      t.integer :chunk_number, null: false
      t.float :duration
      t.integer :size
      t.string :status, default: 'pending'
      t.text :transcription
      t.jsonb :metadata, default: {}
      t.boolean :created_followup_questions, default: false
      t.boolean :is_followup_question, default: false

      t.timestamps
    end

    add_index :audio_chunks, [:session_id, :chunk_number], unique: true
    add_index :audio_chunks, :session_id
    add_index :audio_chunks, :status
  end
end 