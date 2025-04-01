class AddStatusToQuestions < ActiveRecord::Migration[7.1]
  def change
    add_column :questions, :status, :string, default: 'pending'
  end
end 