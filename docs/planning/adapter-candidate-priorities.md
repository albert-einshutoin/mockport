# Adapter Candidate Priorities

Mockport is a Docker-first service emulator for AI-safe, secret-free integration testing.

This document lists external SaaS / Web API / business API adapter candidates by priority tier.

> **Status: exploratory planning document.** The tiers and ordering below are inputs for issue and release planning, not commitments or claims of current support. For the authoritative current surface, see the [support matrix](../site/support-matrix.md). For committed near-term work, see the [project roadmap](../../ROADMAP.md).

Candidate adapter names are provisional planning identifiers. A candidate may become a capability of an existing provider adapter instead of a separate runtime adapter; for example, the current `line` adapter already spans Messaging API, LINE Login, and LINE Pay workflows.

## Scope

Mockport intentionally avoids competing directly with infrastructure emulators such as LocalStack, Flox-style development environments, full database emulators, Kubernetes tooling, Terraform tooling, and complete object-storage replacements.

### Included

- External SaaS APIs
- Payment APIs
- Auth / OAuth / OIDC providers
- Messaging / email / notification APIs
- AI / LLM / voice APIs
- CRM / CS / sales SaaS APIs
- Productivity / no-code APIs
- E-commerce APIs
- Analytics / marketing APIs
- Maps / search / content APIs
- Social / media APIs
- Finance / accounting / booking / HR / legal APIs

### Excluded

- Generic cloud infrastructure emulation
- AWS / GCP / Azure broad emulation
- Kubernetes / Terraform / Docker / CI runner emulation
- Full databases such as PostgreSQL, MySQL, MongoDB, Redis, Kafka
- Full S3-compatible storage replacements
- Generic mock server positioning

---

## Tier Definitions

| Tier | Meaning |
|---|---|
| S | Core Mockport adapters. These should define the product identity and be prioritized first. |
| A | High-value adapters for SaaS, AI, auth, payment, messaging, and modern web app development. |
| B | Strong ecosystem adapters once adapter families and reports are stable. |
| C | Valuable later-stage adapters with broader scope, higher maintenance, or more vertical-specific behavior. |
| D | Long-term, industry-specific, regulation-heavy, or complex adapters. |

---

## Tier S

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| S | Payment | Stripe | `stripe` | Checkout, PaymentIntent, webhook, card_declined, rate limit |
| S | AI API | OpenAI | `openai` | chat, responses, streaming, embeddings, quota, rate limit |
| S | Auth | GitHub OAuth | `github-oauth` | authorize, token, user, invalid code |
| S | Messaging | Slack | `slack` | auth.test, chat.postMessage, rate limit |
| S | Auth | Google OAuth / OIDC | `google-oauth` | authorize, token, userinfo, JWKS, PKCE |
| S | Messaging | LINE Messaging API | `line-messaging` | push, reply, webhook signature, delivery failure |
| S | Auth | LINE Login | `line-login` | authorize, token, profile, state mismatch |
| S | Email | SendGrid | `sendgrid` | mail send, template error, bounce webhook |
| S | Messaging | Discord | `discord` | webhook, bot message, rate limit |
| S | Payment | PayPal | `paypal` | order create, capture, refund, webhook |
| S | SMS / Voice | Twilio | `twilio` | SMS, OTP, callback, delivery status |
| S | Email | Resend | `resend` | email send, domain error, rate limit |
| S | Auth | Auth0 | `auth0` | OIDC, token, userinfo, JWKS |
| S | Auth | Clerk | `clerk` | sessions, users, webhooks |
| S | BaaS Auth | Firebase Auth | `firebase-auth` | signIn, verify token, user, auth error |
| S | BaaS Auth | Supabase Auth | `supabase-auth` | signup, login, refresh, magic link |
| S | AI API | Anthropic | `anthropic` | messages, streaming, context length, rate limit |
| S | AI API | Gemini | `gemini` | generateContent, streaming, safety block |
| S | Core | Generic Webhook | `webhook-generic` | HMAC signing, replay, delayed delivery |
| S | Core | OAuth / OIDC Core | `oauth-core` | reusable OAuth engine for adapters |

---

## Tier A

### Auth / Identity

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| A | Auth | AWS Cognito | `cognito-auth` | hosted UI, token, JWKS, userinfo |
| A | Auth | Microsoft Entra ID | `microsoft-entra` | OIDC, Graph token, tenant error |
| A | Auth | Okta | `okta` | OIDC, SAML-like flow, userinfo |
| A | Auth | Apple Sign In | `apple-signin` | authorize, token, identity token |
| A | Auth | WorkOS | `workos` | SSO, Directory Sync, organization |
| A | Auth | Stytch | `stytch` | passwordless, OTP, session |
| A | Auth | Descope | `descope` | OTP, magic link, session |
| A | Auth | FusionAuth | `fusionauth` | login, token, user |
| A | Auth | Keycloak-like | `keycloak-like` | OIDC local compatibility |
| A | Auth | Hanko | `hanko` | passkey, session, user |

### AI / LLM / Voice APIs

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| A | AI API | Mistral | `mistral` | chat, embeddings, rate limit |
| A | AI API | Groq | `groq` | OpenAI-compatible chat |
| A | AI API | Cohere | `cohere` | generate, embed, rerank |
| A | AI API | xAI | `xai` | chat, streaming |
| A | AI API | Perplexity | `perplexity` | search/chat response |
| A | AI API | Together AI | `together` | chat, image, embeddings |
| A | AI API | Fireworks AI | `fireworks` | inference, streaming |
| A | AI API | Replicate | `replicate` | prediction create/status/webhook |
| A | AI API | Hugging Face Inference | `huggingface` | inference, model loading error |
| A | AI Voice | ElevenLabs | `elevenlabs` | TTS, quota, voice not found |
| A | AI Voice | Deepgram | `deepgram` | speech-to-text, webhook |
| A | AI Voice | AssemblyAI | `assemblyai` | transcription job, webhook |
| A | AI Assistant | Pinecone Assistant | `pinecone-assistant` | assistant chat, file indexing |

### Email / Messaging / Notification

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| A | Email | Mailgun | `mailgun` | send, domain error, webhook |
| A | Email | Postmark | `postmark` | send, bounce, delivery |
| A | Email | Mailchimp Transactional / Mandrill | `mandrill` | send-template, reject |
| A | Email | Brevo | `brevo` | transactional email |
| A | Email | Amazon SES API-like | `ses-lite` | sendEmail, bounce webhook |
| A | Push | OneSignal | `onesignal` | push, delivery failed |
| A | Push | Firebase Cloud Messaging | `firebase-fcm` | push notification, invalid token |
| A | Push | Expo Push | `expo-push` | push ticket, receipt |
| A | Push | Pusher Beams | `pusher-beams` | push publish |
| A | Notification | Novu | `novu` | notification workflow trigger |
| A | Notification | Courier | `courier` | multi-channel notification |
| A | Notification | Knock | `knock` | notification workflow |
| A | Messaging | Telegram Bot API | `telegram-bot` | sendMessage, webhook |
| A | Messaging | WhatsApp Business | `whatsapp-business` | template message, webhook |

### Payment / Billing

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| A | Payment | Square | `square` | payment, refund, webhook |
| A | Payment | Adyen | `adyen` | payments, capture, webhook |
| A | Payment | Paddle | `paddle` | checkout, subscription, webhook |
| A | Payment | Lemon Squeezy | `lemonsqueezy` | checkout, license, webhook |
| A | Billing | Chargebee | `chargebee` | subscription, invoice |
| A | Billing | Recurly | `recurly` | subscription billing |
| A | Payment | Braintree | `braintree` | transaction, webhook |
| A | Payment | PAY.JP | `payjp` | charge, customer, webhook |
| A | Payment | GMO Payment Gateway | `gmo-payment` | authorize, capture, callback |
| A | Payment | KOMOJU | `komoju` | payment, webhook |
| A | Payment | Razorpay | `razorpay` | order, payment, webhook |
| A | Payment | Mercado Pago | `mercadopago` | payment, notification |
| A | Payment | Mollie | `mollie` | payment, refund, webhook |
| A | Payment / Transfer | Wise Platform | `wise` | transfer, quote, status |

---

## Tier B

### CRM / Customer Support / Sales SaaS

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | CRM | HubSpot | `hubspot` | contact, deal, webhook |
| B | CRM | Salesforce | `salesforce` | lead, account, OAuth, REST |
| B | CRM | Zoho CRM | `zoho-crm` | lead, contact, OAuth |
| B | CRM | Pipedrive | `pipedrive` | deal, person, activity |
| B | CRM | Close | `close` | lead, activity |
| B | CRM | Copper | `copper` | lead, company |
| B | Customer Support | Zendesk | `zendesk` | ticket, user, webhook |
| B | Customer Support | Intercom | `intercom` | contact, conversation |
| B | Customer Support | Freshdesk | `freshdesk` | ticket, contact |
| B | Customer Support | Help Scout | `helpscout` | conversation |
| B | Customer Support | Crisp | `crisp` | conversation, website |
| B | Customer Support | Front | `front` | message, inbox |
| B | Customer Support | Gorgias | `gorgias` | ecommerce support |
| B | Dev / CS | Linear | `linear` | issue, webhook |
| B | Dev / CS | Jira Cloud | `jira` | issue, project, webhook |

### Productivity / No-code / Office APIs

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | Productivity | Notion | `notion` | database, page, block |
| B | Productivity | Airtable | `airtable` | records, bases |
| B | Productivity | Google Sheets | `google-sheets` | append, update, read |
| B | Productivity | Google Drive | `google-drive` | upload, list, permissions |
| B | Productivity | Google Calendar | `google-calendar` | event create/update |
| B | Productivity | Gmail API | `gmail` | send, draft, label |
| B | Productivity | Microsoft Graph | `microsoft-graph` | mail, calendar, files |
| B | Productivity | Dropbox | `dropbox` | file upload/list |
| B | Productivity | Box | `box` | file upload/list |
| B | Productivity | OneDrive | `onedrive` | file upload/list |
| B | Productivity | Confluence | `confluence` | page, space |
| B | Productivity | Trello | `trello` | card, board, webhook |
| B | Productivity | Asana | `asana` | task, project |
| B | Productivity | ClickUp | `clickup` | task, list |
| B | Productivity | monday.com | `monday` | board, item |
| B | Productivity | Coda | `coda` | doc, table |
| B | Productivity | Smartsheet | `smartsheet` | sheet, row |
| B | Forms | Typeform | `typeform` | form response webhook |
| B | Forms | Jotform | `jotform` | submission webhook |
| B | Legal / Productivity | DocuSign | `docusign` | envelope, signing webhook |

### E-commerce / Commerce SaaS

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | E-commerce | Shopify | `shopify` | product, order, webhook |
| B | E-commerce | WooCommerce | `woocommerce` | order, product, webhook |
| B | E-commerce | BigCommerce | `bigcommerce` | catalog, order |
| B | E-commerce | Magento / Adobe Commerce | `magento` | order, customer |
| B | E-commerce | commercetools | `commercetools` | cart, order |
| B | E-commerce | Saleor | `saleor` | GraphQL commerce |
| B | E-commerce | Medusa | `medusa` | cart, order |
| B | E-commerce | Spree Commerce | `spree` | REST order/product |
| B | E-commerce | BASE | `base` | item, order |
| B | E-commerce | STORES | `stores` | order, item |
| B | E-commerce | EC-CUBE | `ec-cube` | order, product |
| B | E-commerce | Rakuten API | `rakuten` | item/search/order-like workflows |
| B | E-commerce | Amazon SP-API | `amazon-sp-api` | catalog, order, report |
| B | E-commerce | eBay API | `ebay` | listing, order |
| B | E-commerce | Etsy API | `etsy` | listing, order |
| B | E-commerce | TikTok Shop | `tiktok-shop` | product, order |
| B | E-commerce | Shopee Open API | `shopee` | order, logistics |
| B | E-commerce | Lazada Open Platform | `lazada` | order, product |

### Marketing / Analytics / Feature Flags

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | Error Monitoring | Sentry | `sentry` | event ingest |
| B | Monitoring | Datadog | `datadog` | logs, metrics |
| B | Analytics | Segment | `segment` | track, identify |
| B | Analytics | Mixpanel | `mixpanel` | event track |
| B | Analytics | Amplitude | `amplitude` | event track |
| B | Analytics | PostHog | `posthog` | capture, identify |
| B | Analytics | GA4 Measurement Protocol | `ga4` | event collect |
| B | Analytics | Plausible Events API | `plausible` | event |
| B | Analytics | RudderStack | `rudderstack` | track |
| B | Marketing | Customer.io | `customerio` | event, email trigger |
| B | Marketing | Braze | `braze` | user track, campaign |
| B | Marketing | Iterable | `iterable` | event, email |
| B | Marketing | Klaviyo | `klaviyo` | profile, event |
| B | Marketing | Mailchimp Marketing | `mailchimp` | list, campaign |
| B | Marketing | ActiveCampaign | `activecampaign` | contact, automation |
| B | Marketing | Marketo | `marketo` | lead, campaign |
| B | Marketing | Pardot / Account Engagement | `pardot` | prospect, form |
| B | Experimentation | Optimizely | `optimizely` | event/decision |
| B | Feature Flags | LaunchDarkly | `launchdarkly` | flag evaluation |
| B | Feature Flags | Statsig | `statsig` | feature gate, event |

### Search / Maps / Content APIs

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | Search | Algolia | `algolia` | search, index, object |
| B | Search | Typesense Cloud | `typesense-cloud` | search |
| B | Search | Meilisearch Cloud | `meilisearch-cloud` | search |
| B | Search | Elastic Cloud Search API | `elastic-cloud-search` | query |
| B | Maps | Google Maps | `google-maps` | geocode, places, directions |
| B | Maps | Mapbox | `mapbox` | geocode, directions |
| B | Maps | HERE Maps | `here-maps` | geocode, route |
| B | Maps | OpenRouteService | `openrouteservice` | route |
| B | Maps | MapTiler | `maptiler` | geocoding/tiles metadata |
| B | Geolocation | IPinfo | `ipinfo` | IP geolocation |
| B | Geolocation | ipapi | `ipapi` | IP geolocation |
| B | Content | Contentful | `contentful` | entries, assets |
| B | Content | Sanity | `sanity` | query, document |
| B | Content | Strapi Cloud API | `strapi-cloud` | content API |
| B | Content | Prismic | `prismic` | documents |
| B | Content | Storyblok | `storyblok` | stories |
| B | Content | Hygraph | `hygraph` | GraphQL content |
| B | Content | WordPress.com API | `wordpress` | posts, media |
| B | Content | microCMS | `microcms` | content fetch/create |

---

## Tier C

### Social / Media APIs

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| C | Social | X API | `x-twitter` | post, user, rate limit |
| C | Social | Meta Graph API | `meta-graph` | pages, posts, webhook |
| C | Social | Instagram Graph API | `instagram-graph` | media, insights |
| C | Social | TikTok API | `tiktok` | video, user |
| C | Social | YouTube Data API | `youtube` | videos, channels |
| C | Social | LinkedIn API | `linkedin` | profile, post |
| C | Social | Pinterest API | `pinterest` | pins, boards |
| C | Social | Twitch API | `twitch` | users, streams |
| C | Social | Reddit API | `reddit` | posts, comments |
| C | Social | Mastodon API | `mastodon` | statuses |
| C | Social | Bluesky AT Protocol | `bluesky` | posts, profiles |
| C | Media | Vimeo API | `vimeo` | video upload/status |
| C | Media | Cloudinary | `cloudinary` | upload, transform |
| C | Media | Imgix | `imgix` | image URL/signing |
| C | Media | Mux | `mux` | video asset, webhook |
| C | Media | Livepeer | `livepeer` | stream, asset |
| C | Media | Daily | `daily` | room, meeting |
| C | Media | Zoom API | `zoom` | meeting, webhook |
| C | Media | Agora | `agora` | token, channel |

### Travel / Booking / Lifestyle APIs

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| C | Booking | Calendly | `calendly` | event, invitee webhook |
| C | Booking | Cal.com | `calcom` | booking, webhook |
| C | Business Profile | Google Business Profile | `google-business-profile` | reviews, locations |
| C | Travel | Amadeus API | `amadeus` | flight search, booking mock |
| C | Travel | Skyscanner API | `skyscanner` | search |
| C | Travel | Booking.com Partner API | `booking` | hotel availability |
| C | Travel | Airbnb-like mock | `airbnb-like` | listing, booking |
| C | Delivery | Uber Direct | `uber-direct` | delivery quote/order |
| C | Delivery | DoorDash Drive | `doordash-drive` | delivery |
| C | Delivery | Wolt Drive | `wolt-drive` | delivery |
| C | Delivery | Shopify Fulfillment | `shopify-fulfillment` | fulfillment callback |
| C | Fitness | Strava API | `strava` | activities, athlete |
| C | Fitness | Fitbit Web API | `fitbit` | user data |
| C | Fitness | Garmin Health API | `garmin-health` | activity webhook |
| C | Healthcare-like Booking | Health booking-like flows | `health-booking-like` | appointment callback |

### Finance / Accounting / Business Ops

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| C | Accounting | QuickBooks | `quickbooks` | invoice, customer |
| C | Accounting | Xero | `xero` | invoice, contact |
| C | Accounting | freee | `freee` | invoice/accounting |
| C | Accounting | Money Forward | `moneyforward` | invoice/accounting |
| C | Tax | Stripe Tax-like | `stripe-tax-like` | tax calculation |
| C | Tax | TaxJar | `taxjar` | tax calculation |
| C | Tax | Avalara | `avalara` | tax calculation |
| C | Invoice | Bill.com | `bill-com` | invoice, payment |
| C | Invoice | Misoca | `misoca` | invoice |
| C | Invoice | MakeLeaps | `makeleaps` | invoice |
| C | Banking | Plaid | `plaid` | link token, accounts |
| C | Banking | Tink | `tink` | account aggregation |
| C | Banking | TrueLayer | `truelayer` | open banking |
| C | Trading | Alpaca | `alpaca` | trading paper API |
| C | Crypto | Coinbase API | `coinbase` | account, order |
| C | Crypto | Binance API | `binance` | order, ticker |
| C | Crypto | Kraken API | `kraken` | order, balance |

### HR / Recruiting / Legal

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| C | HR / ATS | Greenhouse | `greenhouse` | candidate, application |
| C | HR / ATS | Lever | `lever` | candidate, posting |
| C | HR / ATS | Ashby | `ashby` | candidate, job |
| C | HR | Workday | `workday` | worker, job |
| C | HR | BambooHR | `bamboohr` | employee |
| C | HR | Personio | `personio` | employee |
| C | HR | SmartHR | `smarthr` | employee |
| C | HR / ATS | HERP | `herp` | candidate |
| C | HR / ATS | Talentio | `talentio` | candidate |
| C | Legal | Dropbox Sign | `dropbox-sign` | signature request |
| C | Legal | CloudSign | `cloudsign` | document signing |
| C | Legal | GMO Sign | `gmo-sign` | signature |
| C | Legal | LegalForce-like | `legalforce-like` | contract status |

---

## Tier D

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| D | Ads | Google Ads | `google-ads` | campaign, conversion, report |
| D | Ads | Meta Ads | `meta-ads` | campaign, insight |
| D | Ads | TikTok Ads | `tiktok-ads` | campaign, report |
| D | Ads | X Ads | `x-ads` | campaign, report |
| D | Ads | LinkedIn Ads | `linkedin-ads` | campaign, report |
| D | Ads | Yahoo Ads Japan | `yahoo-ads-jp` | campaign, report |
| D | Ads | LINE Ads | `line-ads` | campaign, report |
| D | Shipping | Shippo | `shippo` | label, rate, tracking |
| D | Shipping | EasyPost | `easypost` | label, rate, tracking |
| D | Shipping | FedEx | `fedex` | rate, label, tracking |
| D | Shipping | UPS | `ups` | rate, label, tracking |
| D | Shipping | DHL | `dhl` | rate, label, tracking |
| D | Shipping | Yamato | `yamato` | shipment, tracking |
| D | Shipping | Sagawa | `sagawa` | shipment, tracking |
| D | Shipping | Japan Post | `japan-post` | shipment, tracking |
| D | Real Estate | Zillow | `zillow` | property, estimate |
| D | Real Estate | RentCast | `rentcast` | property, rent estimate |
| D | Real Estate | SUUMO-like | `suumo-like` | listing, inquiry |
| D | Education | Canvas LMS | `canvas-lms` | course, assignment |
| D | Education | Moodle API | `moodle` | course, user |
| D | Education | Google Classroom | `google-classroom` | course, coursework |
| D | Identity / KYC | Stripe Identity | `stripe-identity` | verification session |
| D | Identity / KYC | Persona | `persona` | inquiry, webhook |
| D | Identity / KYC | Onfido | `onfido` | applicant, check |
| D | Identity / KYC | Veriff | `veriff` | verification, webhook |
| D | Crypto / Web3 | Alchemy | `alchemy` | JSON-RPC, token API |
| D | Crypto / Web3 | Infura | `infura` | JSON-RPC |
| D | Crypto / Web3 | Moralis | `moralis` | wallet, token API |
| D | Crypto / Web3 | WalletConnect | `walletconnect` | session, callback |
| D | Data Enrichment | Clearbit | `clearbit` | person/company enrichment |
| D | Data Enrichment | Apollo | `apollo` | contact search |
| D | Data Enrichment | Hunter.io | `hunter` | email finder |
| D | Data Enrichment | People Data Labs | `people-data-labs` | person/company lookup |

---

## Adapter Family Strategy

Mockport should not implement every adapter from scratch. It should build reusable adapter families first.

| Family | Reusable For |
|---|---|
| `oauth-family` | Google, GitHub, Auth0, Clerk, Cognito, Okta, LINE Login, Apple Sign In |
| `webhook-family` | Stripe, PayPal, Shopify, SendGrid, LINE, Slack, GitHub App |
| `message-family` | Slack, Discord, LINE, Telegram, WhatsApp |
| `email-family` | SendGrid, Resend, Mailgun, Postmark, Brevo, SES-like |
| `payment-family` | Stripe, PayPal, Square, Adyen, PAY.JP, KOMOJU |
| `ai-chat-family` | OpenAI, Anthropic, Gemini, Mistral, Groq, xAI |
| `ai-job-family` | Replicate, AssemblyAI, Deepgram, Mux, video/transcription APIs |
| `crm-family` | HubSpot, Salesforce, Zoho, Pipedrive |
| `content-family` | Notion, Airtable, Contentful, Sanity, microCMS |
| `analytics-family` | Segment, Mixpanel, Amplitude, PostHog, GA4 |
| `commerce-family` | Shopify, WooCommerce, BigCommerce, Medusa, Saleor |
| `productivity-family` | Google Sheets, Notion, Airtable, Microsoft Graph, Dropbox |

Recommended family implementation order:

1. `webhook-family`
2. `oauth-family`
3. `message-family`
4. `email-family`
5. `ai-chat-family`
6. `payment-family`
7. `ai-job-family`
8. `crm-family`
9. `content-family`
10. `analytics-family`
11. `commerce-family`
12. `productivity-family`

---

## Top 100 Recommended Implementation Order

| Rank | Adapter | Tier | Domain |
|---:|---|---|---|
| 1 | `stripe` | S | Payment |
| 2 | `openai` | S | AI |
| 3 | `github-oauth` | S | Auth |
| 4 | `slack` | S | Messaging |
| 5 | `google-oauth` | S | Auth |
| 6 | `line-messaging` | S | Messaging |
| 7 | `line-login` | S | Auth |
| 8 | `sendgrid` | S | Email |
| 9 | `discord` | S | Messaging |
| 10 | `paypal` | S | Payment |
| 11 | `twilio` | S | SMS / Voice |
| 12 | `resend` | S | Email |
| 13 | `auth0` | S | Auth |
| 14 | `clerk` | S | Auth |
| 15 | `firebase-auth` | S | Auth |
| 16 | `supabase-auth` | S | Auth |
| 17 | `anthropic` | S | AI |
| 18 | `gemini` | S | AI |
| 19 | `webhook-generic` | S | Core |
| 20 | `oauth-core` | S | Core |
| 21 | `mistral` | A | AI |
| 22 | `groq` | A | AI |
| 23 | `cohere` | A | AI |
| 24 | `xai` | A | AI |
| 25 | `perplexity` | A | AI |
| 26 | `replicate` | A | AI |
| 27 | `huggingface` | A | AI |
| 28 | `elevenlabs` | A | AI / Voice |
| 29 | `deepgram` | A | AI / Voice |
| 30 | `assemblyai` | A | AI / Voice |
| 31 | `mailgun` | A | Email |
| 32 | `postmark` | A | Email |
| 33 | `brevo` | A | Email |
| 34 | `ses-lite` | A | Email |
| 35 | `onesignal` | A | Push |
| 36 | `firebase-fcm` | A | Push |
| 37 | `expo-push` | A | Push |
| 38 | `telegram-bot` | A | Messaging |
| 39 | `whatsapp-business` | A | Messaging |
| 40 | `square` | A | Payment |
| 41 | `adyen` | A | Payment |
| 42 | `paddle` | A | Payment |
| 43 | `lemonsqueezy` | A | Payment |
| 44 | `chargebee` | A | Billing |
| 45 | `recurly` | A | Billing |
| 46 | `payjp` | A | Payment |
| 47 | `gmo-payment` | A | Payment |
| 48 | `komoju` | A | Payment |
| 49 | `razorpay` | A | Payment |
| 50 | `mercadopago` | A | Payment |
| 51 | `cognito-auth` | A | Auth |
| 52 | `microsoft-entra` | A | Auth |
| 53 | `okta` | A | Auth |
| 54 | `apple-signin` | A | Auth |
| 55 | `workos` | A | Auth |
| 56 | `stytch` | A | Auth |
| 57 | `descope` | A | Auth |
| 58 | `github-api` | A | Developer API |
| 59 | `github-app` | A | Developer API |
| 60 | `gitlab` | B | Developer API |
| 61 | `vercel` | B | Developer API |
| 62 | `netlify` | B | Developer API |
| 63 | `cloudflare-api-lite` | B | Developer API |
| 64 | `notion` | B | Productivity |
| 65 | `airtable` | B | Productivity |
| 66 | `google-sheets` | B | Productivity |
| 67 | `google-drive` | B | Productivity |
| 68 | `google-calendar` | B | Productivity |
| 69 | `gmail` | B | Productivity |
| 70 | `microsoft-graph` | B | Productivity |
| 71 | `dropbox` | B | Productivity |
| 72 | `box` | B | Productivity |
| 73 | `hubspot` | B | CRM |
| 74 | `salesforce` | B | CRM |
| 75 | `zoho-crm` | B | CRM |
| 76 | `pipedrive` | B | CRM |
| 77 | `zendesk` | B | Customer Support |
| 78 | `intercom` | B | Customer Support |
| 79 | `freshdesk` | B | Customer Support |
| 80 | `front` | B | Customer Support |
| 81 | `linear` | B | Dev / CS |
| 82 | `jira` | B | Dev / CS |
| 83 | `shopify` | B | E-commerce |
| 84 | `woocommerce` | B | E-commerce |
| 85 | `bigcommerce` | B | E-commerce |
| 86 | `medusa` | B | E-commerce |
| 87 | `saleor` | B | E-commerce |
| 88 | `base` | B | E-commerce |
| 89 | `stores` | B | E-commerce |
| 90 | `amazon-sp-api` | C | E-commerce |
| 91 | `sentry` | B | Analytics |
| 92 | `datadog` | B | Analytics |
| 93 | `segment` | B | Analytics |
| 94 | `mixpanel` | B | Analytics |
| 95 | `amplitude` | B | Analytics |
| 96 | `posthog` | B | Analytics |
| 97 | `ga4` | B | Analytics |
| 98 | `google-maps` | B | Maps |
| 99 | `mapbox` | B | Maps |
| 100 | `algolia` | B | Search |

---

## Exploratory Release Sequence

This sequence is a planning hypothesis. A candidate moves into the committed [project roadmap](../../ROADMAP.md) only after its target workflows, maintenance cost, official-reference evidence, and user demand have been validated.

### Foundation: Stabilize Current Adapters

- `stripe`
- `openai`
- `github-oauth`
- `slack`
- `line`
- `zoho-oauth`

Key work:

- Stripe webhook evidence/reporting
- Slack rate-limit headers
- OAuth common error shapes
- Adapter metadata consistency

### Core Adapter Families

- `webhook-family`
- `oauth-family`
- `message-family`
- `email-family`
- `ai-chat-family`

### First Expansion Wave

- `google-oauth`
- `line-messaging`
- `line-login`
- `sendgrid`
- `discord`
- `paypal`
- `twilio`
- `resend`

### SaaS Developer Wave

- `auth0`
- `clerk`
- `firebase-auth`
- `supabase-auth`
- `anthropic`
- `gemini`
- `mailgun`
- `postmark`

### Business API Wave

- `notion`
- `airtable`
- `google-sheets`
- `hubspot`
- `zendesk`
- `shopify`
- `segment`
- `sentry`

---

## Guiding Principle

Mockport should not become a generic mock server.

It should become a curated external SaaS emulator platform:

- service-specific adapters
- fake local secrets
- URL/env switching
- common scenarios
- webhook/callback support
- compatibility reports
- AI-safe warnings
- no real external calls by default
