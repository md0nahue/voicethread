class CreateQuestions < ActiveRecord::Migration[7.1]
  def change
    create_table :questions do |t|
      t.references :interview_section, null: false, foreign_key: true
      t.references :topic, null: false, foreign_key: true
      t.string :body
      t.decimal :llm_fee, precision: 15, scale: 7
      t.string :url

      t.timestamps
    end
  end
end 