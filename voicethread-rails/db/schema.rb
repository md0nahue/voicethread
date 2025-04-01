# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# This file is the source Rails uses to define your schema when running `bin/rails
# db:schema:load`. When creating a new database, `bin/rails db:schema:load` tends to
# be faster and is potentially less error prone than running all of your
# migrations from scratch. Old migrations may fail to apply correctly if those
# migrations use external dependencies or application code.
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema[7.1].define(version: 2024_03_24_000010) do
  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "audio_chunks", force: :cascade do |t|
    t.string "session_id", null: false
    t.string "s3_key", null: false
    t.integer "chunk_number", null: false
    t.float "duration"
    t.integer "size"
    t.string "status", default: "pending"
    t.text "transcription"
    t.jsonb "metadata", default: {}
    t.boolean "created_followup_questions", default: false
    t.boolean "is_followup_question", default: false
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["session_id", "chunk_number"], name: "index_audio_chunks_on_session_id_and_chunk_number", unique: true
    t.index ["session_id"], name: "index_audio_chunks_on_session_id"
    t.index ["status"], name: "index_audio_chunks_on_status"
  end

  create_table "interview_sections", force: :cascade do |t|
    t.bigint "topic_id", null: false
    t.string "body"
    t.decimal "llm_fee", precision: 15, scale: 7
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["topic_id"], name: "index_interview_sections_on_topic_id"
  end

  create_table "interview_sessions", force: :cascade do |t|
    t.bigint "user_id", null: false
    t.bigint "topic_id", null: false
    t.string "status", default: "in_progress"
    t.jsonb "metadata", default: {}
    t.string "final_audio_url"
    t.float "total_duration"
    t.text "polished_transcription"
    t.text "raw_transcript"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["status"], name: "index_interview_sessions_on_status"
    t.index ["topic_id"], name: "index_interview_sessions_on_topic_id"
    t.index ["user_id"], name: "index_interview_sessions_on_user_id"
  end

  create_table "questions", force: :cascade do |t|
    t.bigint "interview_section_id", null: false
    t.bigint "topic_id", null: false
    t.text "body", null: false
    t.string "url"
    t.float "duration"
    t.string "status", default: "pending"
    t.boolean "is_followup_question", default: false
    t.decimal "llm_fee", precision: 15, scale: 7
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.integer "silence_duration", default: 5000
    t.boolean "was_read_to_user", default: false
    t.index ["interview_section_id"], name: "index_questions_on_interview_section_id"
    t.index ["status"], name: "index_questions_on_status"
    t.index ["topic_id"], name: "index_questions_on_topic_id"
  end

  create_table "topics", force: :cascade do |t|
    t.text "body"
    t.decimal "llm_fee", precision: 15, scale: 7
    t.jsonb "fees", default: []
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
  end

  create_table "users", force: :cascade do |t|
    t.string "email", default: "", null: false
    t.string "encrypted_password", default: "", null: false
    t.string "reset_password_token"
    t.datetime "reset_password_sent_at"
    t.datetime "remember_created_at"
    t.datetime "created_at", null: false
    t.datetime "updated_at", null: false
    t.index ["email"], name: "index_users_on_email", unique: true
    t.index ["reset_password_token"], name: "index_users_on_reset_password_token", unique: true
  end

  add_foreign_key "interview_sections", "topics"
  add_foreign_key "interview_sessions", "topics"
  add_foreign_key "interview_sessions", "users"
  add_foreign_key "questions", "interview_sections"
  add_foreign_key "questions", "topics"
end
