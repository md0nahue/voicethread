class CreateTopics < ActiveRecord::Migration[7.1]
  def change
    create_table :topics do |t|
      t.text :body
      t.decimal :llm_fee, precision: 15, scale: 7
      t.jsonb :fees, default: []

      t.timestamps
    end
  end
end 