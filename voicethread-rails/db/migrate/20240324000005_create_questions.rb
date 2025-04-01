class CreateQuestions < ActiveRecord::Migration[7.1]
  def change
    create_table :questions do |t|
      t.references :interview_section, null: false, foreign_key: true
      t.references :topic, null: false, foreign_key: true
      t.text :body, null: false
      t.string :url
      t.float :duration
      t.string :status, default: 'pending'
      t.boolean :is_followup_question, default: false

      t.timestamps
    end

    add_index :questions, :status
  end
end 