# VoiceThread Rails Application

A Rails application for AI-powered interview practice with Stripe subscriptions and AWS services integration.

## Prerequisites

- Ruby 3.1.2
- PostgreSQL
- Node.js (for asset compilation)
- AWS Account
- Stripe Account

## Web Services Setup

### 1. AWS Services Setup

#### AWS SES (Simple Email Service)
1. Go to AWS Console → SES
2. Verify your domain or email address
3. Create a configuration set for tracking and analytics
4. Get your AWS credentials (Access Key ID and Secret Access Key)
5. Add the following to your `.env` file:
   ```
   AWS_ACCESS_KEY_ID=your_access_key
   AWS_SECRET_ACCESS_KEY=your_secret_key
   AWS_REGION=your_region
   AWS_SES_FROM_EMAIL=your_verified_email
   AWS_SES_REPLY_TO_EMAIL=your_reply_to_email
   AWS_SES_CONFIGURATION_SET=your_configuration_set
   ```

#### AWS S3 (Simple Storage Service)
1. Go to AWS Console → S3
2. Create a new bucket
3. Configure CORS for your bucket:
   ```json
   {
     "CORSRules": [
       {
         "AllowedHeaders": ["*"],
         "AllowedMethods": ["GET", "POST", "PUT", "DELETE"],
         "AllowedOrigins": ["*"],
         "ExposeHeaders": []
       }
     ]
   }
   ```
4. Add the following to your `.env` file:
   ```
   AWS_S3_BUCKET=your_bucket_name
   AWS_S3_REGION=your_region
   ```

### 2. Stripe Setup

1. Go to [Stripe Dashboard](https://dashboard.stripe.com)
2. Create a new product (e.g., "VoiceThread Pro")
3. Create a monthly subscription price with a 7-day free trial
4. Get your API keys
5. Add the following to your `.env` file:
   ```
   STRIPE_SECRET_KEY=your_stripe_secret_key
   STRIPE_PUBLISHABLE_KEY=your_stripe_publishable_key
   STRIPE_PRICE_ID=your_stripe_price_id
   ```

### 3. Application Setup

1. Clone the repository
2. Install dependencies:
   ```bash
   bundle install
   ```

3. Set up the database:
   ```bash
   rails db:create db:migrate
   ```

4. Start the development server:
   ```bash
   ./bin/dev
   ```

## Environment Variables

Create a `.env` file in the root directory with the following variables:

```bash
# Database
DATABASE_URL=postgresql://localhost/voicethread_development

# AWS Configuration
AWS_ACCESS_KEY_ID=your_access_key
AWS_SECRET_ACCESS_KEY=your_secret_key
AWS_REGION=your_region
AWS_SES_FROM_EMAIL=your_verified_email
AWS_SES_REPLY_TO_EMAIL=your_reply_to_email
AWS_SES_CONFIGURATION_SET=your_configuration_set
AWS_S3_BUCKET=your_bucket_name
AWS_S3_REGION=your_region

# Stripe Configuration
STRIPE_SECRET_KEY=your_stripe_secret_key
STRIPE_PUBLISHABLE_KEY=your_stripe_publishable_key
STRIPE_PRICE_ID=your_stripe_price_id

# Application Configuration
APP_HOST=your_domain.com
RAILS_MASTER_KEY=your_master_key
```

## Development Workflow

1. The application uses Devise for authentication with email confirmation
2. Users must confirm their email before accessing the application
3. After confirmation, users can subscribe to the service
4. The subscription includes a 7-day free trial
5. Users can create and manage interview topics after subscribing

## Testing

Run the test suite:
```bash
rails test
```

## Deployment

1. Set up your production environment variables
2. Configure your production database
3. Set up SSL certificates
4. Deploy using your preferred method (e.g., Heroku, AWS Elastic Beanstalk)

## Security Considerations

1. Never commit `.env` files or sensitive credentials
2. Use environment variables for all sensitive data
3. Enable SSL in production
4. Regularly rotate AWS and Stripe API keys
5. Monitor AWS SES sending limits and reputation

## Support

For support, please open an issue in the repository or contact the development team.
