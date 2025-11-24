# Microsoft Teams Integration

The TMS allows you to integrate with Microsoft Teams to receive notifications about new chat messages directly in your Teams channels.

## Prerequisites

- A Microsoft Teams account.
- A team and channel where you want to receive notifications.
- Permissions to add connectors to the channel.

## Setup Instructions

### 1. Create an Incoming Webhook in Microsoft Teams

1.  Navigate to the channel where you want to receive notifications.
2.  Click the **... (More options)** button next to the channel name.
3.  Select **Connectors**.
4.  Search for **Incoming Webhook** and click **Configure**.
5.  Provide a name for the webhook (e.g., "TMS Notifications").
6.  (Optional) Upload an image for the webhook avatar.
7.  Click **Create**.
8.  Copy the generated **Webhook URL**. You will need this for the TMS configuration.

### 2. Configure Integration in TMS

1.  Log in to your TMS dashboard.
2.  Navigate to **Settings > Integrations**.
3.  Find **Microsoft Teams** in the list of available integrations.
4.  Click **Connect** or **Configure**.
5.  Enter a name for the integration (e.g., "Support Channel").
6.  Paste the **Webhook URL** you copied from Teams into the configuration field.
7.  Click **Save**.

## How it Works

Once configured, any new message sent by a visitor in the chat widget will be forwarded to the configured Microsoft Teams channel. The message will include:

-   **Sender Name**: The name of the visitor.
-   **Session ID**: The ID of the chat session.
-   **Message Content**: The text of the message.

## Troubleshooting

-   **No messages received**: Ensure the Webhook URL is correct and the integration is active in TMS. Check the TMS logs for any error messages related to Microsoft Teams.
-   **Error 400/401/403**: This usually indicates an issue with the Webhook URL or permissions in Teams. Try regenerating the URL.
