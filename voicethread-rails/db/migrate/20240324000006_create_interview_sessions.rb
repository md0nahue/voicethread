class CreateInterviewSessions < ActiveRecord::Migration[7.1]
  def change
    create_table :interview_sessions do |t|
      t.references :user, null: false, foreign_key: true
      t.references :topic, null: false, foreign_key: true
      t.string :status, default: 'in_progress'
      t.jsonb :metadata, default: {}
      t.string :final_audio_url
      t.float :total_duration
      t.text :polished_transcription
      t.text :raw_transcript

      t.timestamps
    end

    add_index :interview_sessions, :status
  end
end 