class AddSilenceDurationToQuestions < ActiveRecord::Migration[7.1]
  def change
    add_column :questions, :silence_duration, :integer, default: 5000 # Default 5 seconds of silence
  end
end 