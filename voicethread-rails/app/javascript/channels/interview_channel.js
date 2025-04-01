import consumer from "./consumer"

const InterviewChannel = {
  subscription: null,

  connect() {
    this.subscription = consumer.subscriptions.create("InterviewChannel", {
      connected() {
        console.log("Connected to InterviewChannel")
      },

      disconnected() {
        console.log("Disconnected from InterviewChannel")
      },

      received(data) {
        if (data.type === 'interview_questions') {
          // Forward the questions data to the Golang server's WebSocket
          const golangWs = window.golangWebSocket // Assuming this is your Golang WebSocket connection
          if (golangWs && golangWs.readyState === WebSocket.OPEN) {
            golangWs.send(JSON.stringify({
              type: 'interview_questions',
              data: data.data
            }))
          }
        }
      }
    })
  },

  requestQuestions(topicId) {
    if (this.subscription) {
      this.subscription.perform('request_questions', {
        topic_id: topicId
      })
    }
  },

  disconnect() {
    if (this.subscription) {
      this.subscription.unsubscribe()
      this.subscription = null
    }
  }
}

export default InterviewChannel 