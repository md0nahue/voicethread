class AddWasReadToUserToQuestions < ActiveRecord::Migration[7.1]
  def change
    add_column :questions, :was_read_to_user, :boolean, default: false
  end
end 