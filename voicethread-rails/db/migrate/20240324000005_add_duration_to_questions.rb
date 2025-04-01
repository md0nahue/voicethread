class AddDurationToQuestions < ActiveRecord::Migration[7.1]
  def change
    add_column :questions, :duration, :integer # Duration in milliseconds
  end
end 