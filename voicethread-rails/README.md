# VoiceThread

VoiceThread is an AI-powered interview platform that helps users record their life stories through guided conversations. The platform uses AI to generate relevant questions and AWS Polly to read them aloud, creating a natural interview experience.

## Features

- **AI-Generated Questions**: Creates personalized interview questions based on the topic
- **Text-to-Speech**: Questions are read aloud using AWS Polly
- **Smart Interview Flow**: Automatically detects silence and plays the next question
- **User Authentication**: Secure user accounts with email verification
- **Subscription Management**: Stripe integration for premium features
- **Audio Storage**: Secure storage of generated audio files in AWS S3

## Prerequisites

- Ruby 3.1.2
- PostgreSQL
- Redis (for background jobs)
- FFmpeg (for audio processing)
- Node.js (for asset compilation)

## Cloud Services Required

### 1. AWS Services
- **AWS S3**: For storing generated audio files
- **AWS Polly**: For text-to-speech conversion
- **AWS SES**: For sending emails

### 2. OpenAI
- **OpenAI API**: For generating interview questions

### 3. Stripe
- **Stripe API**: For handling subscriptions

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```bash
# Database
DATABASE_URL=postgresql://username:password@localhost:5432/voicethread_development

# AWS Configuration
AWS_ACCESS_KEY_ID=your_aws_access_key
AWS_SECRET_ACCESS_KEY=your_aws_secret_key
AWS_REGION=your_aws_region
AWS_BUCKET=your_s3_bucket_name
AWS_SES_FROM_EMAIL=your_verified_ses_email

# OpenAI
OPENAI_API_KEY=your_openai_api_key

# Stripe
STRIPE_PUBLISHABLE_KEY=your_stripe_publishable_key
STRIPE_SECRET_KEY=your_stripe_secret_key
STRIPE_WEBHOOK_SECRET=your_stripe_webhook_secret

# Redis
REDIS_URL=redis://localhost:6379/1
```

## Setup Instructions

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd voicethread-rails
   ```

2. **Install dependencies**
   ```bash
   bundle install
   ```

3. **Setup database**
   ```bash
   rails db:create db:migrate
   ```

4. **Install JavaScript dependencies**
   ```bash
   yarn install
   ```

5. **Start Redis**
   ```bash
   redis-server
   ```

6. **Start Sidekiq**
   ```bash
   bundle exec sidekiq
   ```

7. **Start the Rails server**
   ```bash
   ./bin/dev
   ```

## Cloud Service Setup

### AWS Setup

1. **Create an S3 bucket**
   - Create a new bucket in AWS S3
   - Enable public access for the bucket
   - Configure CORS for the bucket:
   ```json
   {
       "CORSRules": [
           {
               "AllowedHeaders": ["*"],
               "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
               "AllowedOrigins": ["*"],
               "ExposeHeaders": []
           }
       ]
   }
   ```

2. **Configure AWS Polly**
   - Create an IAM user with Polly access
   - Note down the access key and secret key

3. **Setup AWS SES**
   - Verify your email domain in SES
   - Request production access if needed
   - Configure email sending limits

### OpenAI Setup

1. Create an OpenAI account
2. Generate an API key
3. Add the key to your environment variables

### Stripe Setup

1. Create a Stripe account
2. Get your API keys
3. Configure webhook endpoints:
   - URL: `https://your-domain.com/webhooks/stripe`
   - Events to listen for:
     - `customer.subscription.created`
     - `customer.subscription.updated`
     - `customer.subscription.deleted`

## Development Workflow

1. **Create a new topic**
   - Navigate to `/topics/new`
   - Enter a topic description
   - The system will automatically generate questions

2. **Start an interview**
   - Select a topic
   - The system will play questions during silence
   - Record your responses

3. **View recordings**
   - Access your recordings in the dashboard
   - Download or share as needed

## Testing

```bash
# Run tests
rails test

# Run with coverage
COVERAGE=true rails test
```

## Deployment

The application is designed to be deployed on any cloud platform that supports Ruby on Rails applications. Recommended platforms:

- Heroku
- AWS Elastic Beanstalk
- DigitalOcean App Platform

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
