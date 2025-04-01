// Configure your import map in config/importmap.rb. Read more: https://github.com/rails/importmap-rails
import "@hotwired/turbo-rails"
import "controllers"
import "channels"
import InterviewChannel from "./channels/interview_channel"

// Initialize the interview channel when the page loads
document.addEventListener("turbo:load", () => {
  InterviewChannel.connect()
})

// Clean up when the page is unloaded
document.addEventListener("turbo:before-cache", () => {
  InterviewChannel.disconnect()
})
