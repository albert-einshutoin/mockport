# Adapter Candidate Priorities

Mockport currently emulates six provider families: Stripe, OpenAI, GitHub OAuth, Slack, LINE, and Zoho OAuth. This document helps maintainers decide what to build after that foundation is stable.

It is a planning catalog, not a support matrix or a release schedule.

> **Status: exploratory planning document.** The tiers and ordering below are inputs for issue and release planning, not commitments or claims of current support. For the authoritative current surface, see the [support matrix](../site/support-matrix.md). For committed near-term work, see the [project roadmap](../../ROADMAP.md).

Candidate adapter names are provisional planning identifiers. A candidate may become a capability of an existing provider adapter instead of a separate runtime adapter; for example, the current `line` adapter already spans Messaging API, LINE Login, and LINE Pay workflows.

## Start Here

Use this document in this order:

1. Confirm that the proposed API fits the [scope](#scope).
2. Apply the [selection gates](#selection-gates) before assigning a tier.
3. Use the [tier definitions](#tier-definitions) to compare candidates.
4. Check the [adapter family strategy](#adapter-family-strategy) for reusable behavior, without creating premature shared abstractions.
5. Move a candidate into the [project roadmap](../../ROADMAP.md) only after its first supported workflow and evidence plan are concrete.

## Current Baseline

These adapters are already built in. Work on them is stabilization or compatibility expansion, not new-adapter delivery.

| Provider family | Runtime adapter | Current focus |
|---|---|---|
| Stripe | `stripe` | Selected payment, checkout, state, and webhook workflows |
| OpenAI | `openai` | Selected models, chat, Responses, streaming, and embedding workflows |
| GitHub OAuth | `github-oauth` | Authorization, token exchange, and profile workflows |
| Slack | `slack` | Authentication, conversations, and message workflows |
| LINE | `line` | Messaging API, LINE Login, and LINE Pay workflows |
| Zoho OAuth | `zoho-oauth` | Authorization, token exchange, and user information workflows |

The [support matrix](../site/support-matrix.md), individual [adapter specifications](../adapters/), runtime metadata, and compatibility reports remain the source of truth for exact coverage.

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

## Selection Gates

A candidate should enter the committed roadmap only when all of these questions have useful answers:

1. **User workflow:** Which integration test becomes possible, and who needs it?
2. **Bounded contract:** Which endpoints, errors, state transitions, callbacks, or webhooks form the first supported workflow?
3. **Local determinism:** Can the workflow run without real credentials, external network calls, or provider-owned state?
4. **Evidence:** Can official documentation, sanitized fixtures, SDK contract tests, and known-gap reporting support the compatibility claim?
5. **Maintenance cost:** Can Mockport track provider drift without promising a full clone?
6. **Safety:** Can examples and defaults remain secret-free and safe for AI agents and CI?

Popularity alone is not enough. Prefer a smaller adapter with a clear, testable workflow over a famous provider with an unbounded API surface.

---

## Tier Definitions

| Tier | Meaning |
|---|---|
| S | Product-defining provider families and shared capabilities. Includes the current foundation and the strongest next candidates. |
| A | High-value candidates with common web application workflows and a plausible bounded first contract. |
| B | Strong ecosystem candidates to consider after the relevant adapter pattern and evidence pipeline are proven. |
| C | Later candidates with broader scope, higher maintenance cost, or more vertical-specific behavior. |
| D | Long-term candidates that are industry-specific, regulation-heavy, operationally complex, or difficult to emulate safely. |

Tiers express strategic value, not implementation readiness. A lower-tier candidate with a strong user request and a tightly bounded workflow may be scheduled before a higher-tier candidate that lacks evidence or maintainership.

---

## Tier S

### Current Foundation

These provider families define the current product. Their exact supported workflows remain intentionally narrower than the full provider APIs.

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| S | Payment | Stripe | `stripe` | Checkout, PaymentIntent, webhook, card_declined, rate limit |
| S | AI API | OpenAI | `openai` | chat, responses, streaming, embeddings, quota, rate limit |
| S | Auth | GitHub OAuth | `github-oauth` | authorize, token, user, invalid code |
| S | Messaging | Slack | `slack` | auth.test, chat.postMessage, rate limit |
| S | Platform APIs | LINE | `line` | Messaging API, Login, Pay, webhook signature, delivery failure |
| S | Auth | Zoho OAuth | `zoho-oauth` | authorize, token, userinfo, invalid code |

### Strongest Next Candidates

These candidates fit Mockport's product direction, but each still needs a bounded first workflow and evidence plan before roadmap commitment.

| Tier | Domain | Service | Planning Name | Main Mock Targets |
|---|---|---|---|---|
| S | Auth | Google OAuth / OIDC | `google-oauth` | authorize, token, userinfo, JWKS, PKCE |
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

### Shared Capability Candidates

These are reusable capabilities, not standalone provider adapters.

| Tier | Capability | Planning Name | Intended Use |
|---|---|---|---|
| S | Generic Webhook | `webhook-generic` | HMAC signing, replay, and delayed delivery across webhook-based adapters |
| S | OAuth / OIDC Core | `oauth-core` | Tested protocol primitives shared by compatible OAuth adapters |

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

### Developer Platforms

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| A | Developer API | GitHub API | `github-api` | repository, issue, pull request, rate limit |
| A | Developer API | GitHub App | `github-app` | installation token, webhook, permissions |

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

### Developer and Deployment Platforms

| Tier | Domain | Service | Adapter Name | Main Mock Targets |
|---|---|---|---|---|
| B | Developer API | GitLab | `gitlab` | project, merge request, pipeline, webhook |
| B | Deployment | Vercel | `vercel` | deployment, status, webhook |
| B | Deployment | Netlify | `netlify` | site, deployment, build hook |
| B | Edge Platform | Cloudflare API | `cloudflare-api-lite` | zone, DNS record, cache purge |

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

Adapter families are a planning lens for repeated protocol and workflow behavior. They are **not** instructions to create a shared package before provider-specific contracts exist.

Implement the first provider contract locally. Extract a shared primitive only when at least two adapters have tested, genuinely identical invariants and the extraction preserves provider-specific errors, validation, and state behavior. Follow the [adapter helper policy](../adapter-helper-policy.md) when deciding whether code should remain local.

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

Recommended order for proving reusable patterns:

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

This order describes where reuse is likely to pay off. It does not override candidate demand, the selection gates, or the committed roadmap.

---

## Appendix A: Ranked Candidate Backlog

This ranking is a coarse comparison aid, not a build queue. Re-evaluate a row with the selection gates before creating an issue. The current built-in adapters appear at the top only to anchor product priorities; they are not new implementation work.

| Rank | Adapter | Tier | Domain |
|---:|---|---|---|
| 1 | `stripe` | S | Payment |
| 2 | `openai` | S | AI |
| 3 | `github-oauth` | S | Auth |
| 4 | `slack` | S | Messaging |
| 5 | `line` | S | Platform APIs |
| 6 | `zoho-oauth` | S | Auth |
| 7 | `google-oauth` | S | Auth |
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

## Appendix B: Example Sequencing

This sequence shows one plausible progression from the current foundation. It is not a version plan. A candidate moves into the committed [project roadmap](../../ROADMAP.md) only after its target workflows, maintenance cost, official-reference evidence, and user demand have been validated.

### Foundation: Stabilize Current Adapters

- `stripe`
- `openai`
- `github-oauth`
- `slack`
- `line`
- `zoho-oauth`

Example stabilization work:

- Stripe webhook evidence/reporting
- Slack rate-limit headers
- OAuth common error shapes
- Adapter metadata consistency

### Core Adapter Families

- Prove reusable webhook behavior across existing and new provider contracts.
- Prove reusable OAuth/OIDC behavior without flattening provider-specific errors.
- Keep message, email, and AI response helpers local until multiple tested contracts justify extraction.

### First Expansion Wave

- `google-oauth`
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

## Decision Summary

Mockport should not become a generic mock server.

It should remain a curated external SaaS emulator platform. Every roadmap adapter should provide:

- a service-specific, bounded workflow;
- fake local secrets and safe URL or environment switching;
- deterministic success and failure scenarios;
- webhook or callback behavior where the workflow requires it;
- compatibility evidence and explicit known gaps;
- AI-safe warnings and examples; and
- no real external calls by default.

If a candidate cannot meet those conditions, it should remain in this catalog rather than enter the committed roadmap.
