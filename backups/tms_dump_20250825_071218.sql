--
-- PostgreSQL database dump
--

-- Dumped from database version 15.13 (Debian 15.13-1.pgdg120+1)
-- Dumped by pg_dump version 15.13 (Debian 15.13-1.pgdg120+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: agent_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.agent_status AS ENUM (
    'active',
    'inactive',
    'suspended'
);


ALTER TYPE public.agent_status OWNER TO tms;

--
-- Name: article_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.article_status AS ENUM (
    'draft',
    'published',
    'archived',
    'under_review'
);


ALTER TYPE public.article_status OWNER TO tms;

--
-- Name: article_visibility; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.article_visibility AS ENUM (
    'public',
    'internal',
    'private'
);


ALTER TYPE public.article_visibility OWNER TO tms;

--
-- Name: attachment_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.attachment_type AS ENUM (
    'ticket_attachment',
    'message_attachment',
    'agent_avatar',
    'org_logo',
    'knowledge_article'
);


ALTER TYPE public.attachment_type OWNER TO tms;

--
-- Name: author_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.author_type AS ENUM (
    'agent',
    'customer',
    'system'
);


ALTER TYPE public.author_type OWNER TO tms;

--
-- Name: integration_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.integration_status AS ENUM (
    'active',
    'inactive',
    'error',
    'configuring'
);


ALTER TYPE public.integration_status OWNER TO tms;

--
-- Name: integration_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.integration_type AS ENUM (
    'slack',
    'jira',
    'calendly',
    'zapier',
    'webhook',
    'custom',
    'microsoft_teams',
    'github',
    'linear',
    'asana',
    'trello',
    'monday',
    'notion',
    'airtable',
    'hubspot',
    'salesforce',
    'zendesk',
    'freshdesk',
    'intercom',
    'crisp',
    'discord',
    'telegram',
    'whatsapp',
    'google_drive',
    'dropbox',
    'box',
    'onedrive',
    'aws_s3',
    'azure_storage',
    'google_cloud_storage',
    'stripe',
    'paypal',
    'square',
    'twilio',
    'sendgrid',
    'mailchimp',
    'constant_contact',
    'google_calendar',
    'outlook_calendar',
    'zoom',
    'google_meet',
    'microsoft_teams_meeting',
    'shopify',
    'woocommerce',
    'magento',
    'bigcommerce'
);


ALTER TYPE public.integration_type OWNER TO tms;

--
-- Name: job_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.job_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'cancelled',
    'retrying'
);


ALTER TYPE public.job_status OWNER TO tms;

--
-- Name: job_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.job_type AS ENUM (
    'email_send',
    'email_process',
    'notification_send',
    'integration_sync',
    'webhook_delivery',
    'report_generation',
    'data_export',
    'data_import',
    'cleanup',
    'backup'
);


ALTER TYPE public.job_type OWNER TO tms;

--
-- Name: metric_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.metric_type AS ENUM (
    'counter',
    'gauge',
    'histogram',
    'timer'
);


ALTER TYPE public.metric_type OWNER TO tms;

--
-- Name: notification_channel; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.notification_channel AS ENUM (
    'web',
    'email',
    'slack',
    'sms',
    'push'
);


ALTER TYPE public.notification_channel OWNER TO tms;

--
-- Name: notification_priority; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.notification_priority AS ENUM (
    'low',
    'normal',
    'high',
    'urgent'
);


ALTER TYPE public.notification_priority OWNER TO tms;

--
-- Name: notification_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.notification_type AS ENUM (
    'ticket_assigned',
    'ticket_updated',
    'ticket_escalated',
    'ticket_resolved',
    'message_received',
    'mention_received',
    'sla_warning',
    'sla_breach',
    'system_alert',
    'maintenance_notice',
    'feature_announcement'
);


ALTER TYPE public.notification_type OWNER TO tms;

--
-- Name: oauth_provider; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.oauth_provider AS ENUM (
    'google',
    'microsoft',
    'slack',
    'jira',
    'custom',
    'github',
    'linear',
    'asana',
    'trello',
    'notion',
    'hubspot',
    'salesforce',
    'zendesk',
    'freshdesk',
    'intercom',
    'discord',
    'stripe',
    'shopify',
    'zoom',
    'calendly'
);


ALTER TYPE public.oauth_provider OWNER TO tms;

--
-- Name: project_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.project_status AS ENUM (
    'active',
    'inactive'
);


ALTER TYPE public.project_status OWNER TO tms;

--
-- Name: role_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.role_type AS ENUM (
    'tenant_admin',
    'project_admin',
    'supervisor',
    'agent',
    'read_only'
);


ALTER TYPE public.role_type OWNER TO tms;

--
-- Name: search_entity_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.search_entity_type AS ENUM (
    'ticket',
    'message',
    'customer',
    'agent',
    'organization',
    'knowledge_article'
);


ALTER TYPE public.search_entity_type OWNER TO tms;

--
-- Name: storage_provider; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.storage_provider AS ENUM (
    'minio',
    's3',
    'azure_blob',
    'gcs'
);


ALTER TYPE public.storage_provider OWNER TO tms;

--
-- Name: tenant_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.tenant_status AS ENUM (
    'active',
    'inactive',
    'suspended'
);


ALTER TYPE public.tenant_status OWNER TO tms;

--
-- Name: ticket_priority; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.ticket_priority AS ENUM (
    'low',
    'normal',
    'high',
    'urgent'
);


ALTER TYPE public.ticket_priority OWNER TO tms;

--
-- Name: ticket_source; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.ticket_source AS ENUM (
    'web',
    'email',
    'api',
    'phone',
    'chat'
);


ALTER TYPE public.ticket_source OWNER TO tms;

--
-- Name: ticket_status; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.ticket_status AS ENUM (
    'new',
    'open',
    'pending',
    'resolved',
    'closed'
);


ALTER TYPE public.ticket_status OWNER TO tms;

--
-- Name: ticket_type; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.ticket_type AS ENUM (
    'question',
    'incident',
    'problem',
    'task'
);


ALTER TYPE public.ticket_type OWNER TO tms;

--
-- Name: unauth_scope; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.unauth_scope AS ENUM (
    'view',
    'reply'
);


ALTER TYPE public.unauth_scope OWNER TO tms;

--
-- Name: webhook_event; Type: TYPE; Schema: public; Owner: tms
--

CREATE TYPE public.webhook_event AS ENUM (
    'ticket.created',
    'ticket.updated',
    'ticket.status_changed',
    'message.created',
    'message.updated',
    'agent.assigned',
    'agent.unassigned',
    'escalation.triggered',
    'sla.breached'
);


ALTER TYPE public.webhook_event OWNER TO tms;

--
-- Name: set_ticket_number(); Type: FUNCTION; Schema: public; Owner: tms
--

CREATE FUNCTION public.set_ticket_number() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.number IS NULL THEN
        SELECT COALESCE(MAX(number), 0) + 1 INTO NEW.number
        FROM tickets 
        WHERE tenant_id = NEW.tenant_id;
    END IF;
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.set_ticket_number() OWNER TO tms;

--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: tms
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


ALTER FUNCTION public.update_updated_at_column() OWNER TO tms;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: agent_project_roles; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.agent_project_roles (
    agent_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    role public.role_type NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.agent_project_roles OWNER TO tms;

--
-- Name: agents; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.agents (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    email character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    status public.agent_status DEFAULT 'active'::public.agent_status NOT NULL,
    password_hash text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.agents OWNER TO tms;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.api_keys (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid,
    name character varying(255) NOT NULL,
    key_hash character varying(255) NOT NULL,
    key_prefix character varying(20) NOT NULL,
    scopes jsonb DEFAULT '[]'::jsonb,
    last_used_at timestamp with time zone,
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_by uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.api_keys OWNER TO tms;

--
-- Name: attachments; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    ticket_id uuid NOT NULL,
    message_id uuid,
    blob_key character varying(500) NOT NULL,
    filename character varying(255) NOT NULL,
    content_type character varying(100) NOT NULL,
    size_bytes bigint NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.attachments OWNER TO tms;

--
-- Name: chat_messages; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.chat_messages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    session_id uuid NOT NULL,
    message_type character varying(20) DEFAULT 'text'::character varying,
    content text NOT NULL,
    author_type character varying(20) NOT NULL,
    author_id uuid,
    author_name character varying(255),
    metadata jsonb DEFAULT '{}'::jsonb,
    is_private boolean DEFAULT false,
    read_by_visitor boolean DEFAULT false,
    read_by_agent boolean DEFAULT false,
    read_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT chat_messages_author_type_check CHECK (((author_type)::text = ANY ((ARRAY['visitor'::character varying, 'agent'::character varying, 'system'::character varying, 'ai-agent'::character varying])::text[]))),
    CONSTRAINT chat_messages_message_type_check CHECK (((message_type)::text = ANY ((ARRAY['text'::character varying, 'file'::character varying, 'image'::character varying, 'system'::character varying])::text[])))
);


ALTER TABLE public.chat_messages OWNER TO tms;

--
-- Name: chat_session_participants; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.chat_session_participants (
    session_id uuid NOT NULL,
    agent_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    role character varying(20) DEFAULT 'participant'::character varying,
    joined_at timestamp with time zone DEFAULT now() NOT NULL,
    left_at timestamp with time zone,
    CONSTRAINT chat_session_participants_role_check CHECK (((role)::text = ANY ((ARRAY['primary'::character varying, 'participant'::character varying, 'observer'::character varying])::text[])))
);


ALTER TABLE public.chat_session_participants OWNER TO tms;

--
-- Name: chat_sessions; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.chat_sessions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    widget_id uuid NOT NULL,
    customer_id uuid,
    ticket_id uuid,
    status character varying(20) DEFAULT 'active'::character varying,
    visitor_info jsonb DEFAULT '{}'::jsonb,
    assigned_agent_id uuid,
    assigned_at timestamp with time zone,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    ended_at timestamp with time zone,
    last_activity_at timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    client_session_id character varying(255) NOT NULL,
    CONSTRAINT chat_sessions_status_check CHECK (((status)::text = ANY ((ARRAY['active'::character varying, 'ended'::character varying, 'transferred'::character varying])::text[])))
);


ALTER TABLE public.chat_sessions OWNER TO tms;

--
-- Name: chat_widgets; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.chat_widgets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    domain_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    is_active boolean DEFAULT true,
    primary_color character varying(7) DEFAULT '#2563eb'::character varying,
    secondary_color character varying(7) DEFAULT '#f3f4f6'::character varying,
    "position" character varying(20) DEFAULT 'bottom-right'::character varying,
    welcome_message text DEFAULT 'Hello! How can we help you?'::text,
    offline_message text DEFAULT 'We are currently offline. Please leave a message.'::text,
    auto_open_delay integer DEFAULT 0,
    show_agent_avatars boolean DEFAULT true,
    allow_file_uploads boolean DEFAULT true,
    require_email boolean DEFAULT true,
    business_hours jsonb DEFAULT '{"enabled": false}'::jsonb,
    embed_code text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    widget_shape character varying(50) DEFAULT 'rounded'::character varying,
    chat_bubble_style character varying(50) DEFAULT 'modern'::character varying,
    agent_name character varying(255) DEFAULT 'Support Agent'::character varying,
    agent_avatar_url text,
    use_ai boolean DEFAULT false,
    custom_css text,
    widget_size character varying(20) DEFAULT 'medium'::character varying,
    animation_style character varying(30) DEFAULT 'smooth'::character varying,
    sound_enabled boolean DEFAULT true,
    show_powered_by boolean DEFAULT true,
    custom_greeting text,
    away_message text DEFAULT 'We''re currently away. Leave us a message and we''ll get back to you!'::text,
    background_color character varying(7) DEFAULT '#ffffff'::character varying,
    require_name boolean DEFAULT false NOT NULL,
    CONSTRAINT chat_bubble_style_check CHECK (((chat_bubble_style)::text = ANY ((ARRAY['modern'::character varying, 'classic'::character varying, 'minimal'::character varying, 'bot'::character varying])::text[]))),
    CONSTRAINT chat_widgets_animation_style_check CHECK (((animation_style)::text = ANY ((ARRAY['smooth'::character varying, 'bounce'::character varying, 'fade'::character varying, 'slide'::character varying])::text[]))),
    CONSTRAINT chat_widgets_position_check CHECK ((("position")::text = ANY ((ARRAY['bottom-right'::character varying, 'bottom-left'::character varying])::text[]))),
    CONSTRAINT chat_widgets_widget_shape_check CHECK (((widget_shape)::text = ANY ((ARRAY['rounded'::character varying, 'square'::character varying, 'minimal'::character varying, 'professional'::character varying, 'modern'::character varying, 'classic'::character varying])::text[]))),
    CONSTRAINT chat_widgets_widget_size_check CHECK (((widget_size)::text = ANY ((ARRAY['small'::character varying, 'medium'::character varying, 'large'::character varying])::text[])))
);


ALTER TABLE public.chat_widgets OWNER TO tms;

--
-- Name: COLUMN chat_widgets.widget_shape; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.widget_shape IS 'Visual theme/shape of the chat widget UI';


--
-- Name: COLUMN chat_widgets.chat_bubble_style; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.chat_bubble_style IS 'Style of chat message bubbles';


--
-- Name: COLUMN chat_widgets.agent_name; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.agent_name IS 'Personalized agent name shown to visitors';


--
-- Name: COLUMN chat_widgets.agent_avatar_url; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.agent_avatar_url IS 'URL to agent profile picture';


--
-- Name: COLUMN chat_widgets.use_ai; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.use_ai IS 'Whether AI assistance is enabled for this widget';


--
-- Name: COLUMN chat_widgets.custom_css; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.custom_css IS 'Additional CSS customizations';


--
-- Name: COLUMN chat_widgets.widget_size; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.widget_size IS 'Size variant of the chat widget';


--
-- Name: COLUMN chat_widgets.animation_style; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.animation_style IS 'Animation style for widget interactions';


--
-- Name: COLUMN chat_widgets.sound_enabled; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.sound_enabled IS 'Whether notification sounds are enabled';


--
-- Name: COLUMN chat_widgets.show_powered_by; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.show_powered_by IS 'Whether to show "Powered by" branding';


--
-- Name: COLUMN chat_widgets.custom_greeting; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.custom_greeting IS 'Custom greeting message (overrides welcome_message)';


--
-- Name: COLUMN chat_widgets.away_message; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.chat_widgets.away_message IS 'Message shown when agents are offline';


--
-- Name: customers; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.customers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    email character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    org_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    metadata jsonb
);


ALTER TABLE public.customers OWNER TO tms;

--
-- Name: email_attachments; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_attachments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    filename text NOT NULL,
    content_type text NOT NULL,
    size_bytes integer NOT NULL,
    content_id text,
    is_inline boolean DEFAULT false,
    storage_path text,
    storage_url text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    project_id uuid
);


ALTER TABLE public.email_attachments OWNER TO tms;

--
-- Name: email_connectors; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_connectors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    type text NOT NULL,
    name text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    imap_host text,
    imap_port integer,
    imap_use_tls boolean DEFAULT true,
    imap_username text,
    imap_password_enc bytea,
    imap_folder text DEFAULT 'INBOX'::text,
    imap_seen_strategy text DEFAULT 'mark_seen_after_parse'::text,
    smtp_host text NOT NULL,
    smtp_port integer NOT NULL,
    smtp_use_tls boolean DEFAULT true,
    smtp_username text NOT NULL,
    smtp_password_enc bytea NOT NULL,
    oauth_provider text,
    oauth_account_email text,
    oauth_token_ref uuid,
    dkim_selector text,
    dkim_public_key text,
    dkim_private_key_enc bytea,
    return_path_domain text,
    provider_webhook_secret text,
    last_health jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    project_id uuid NOT NULL,
    is_validated boolean DEFAULT false,
    validation_status text DEFAULT 'pending'::text,
    validation_error text,
    last_validation_at timestamp with time zone,
    CONSTRAINT email_connectors_imap_seen_strategy_check CHECK ((imap_seen_strategy = ANY (ARRAY['mark_seen_after_parse'::text, 'never'::text, 'immediate'::text]))),
    CONSTRAINT email_connectors_oauth_provider_check CHECK ((oauth_provider = ANY (ARRAY['google'::text, 'microsoft'::text]))),
    CONSTRAINT email_connectors_type_check CHECK ((type = ANY (ARRAY['inbound_imap'::text, 'outbound_smtp'::text, 'outbound_provider'::text]))),
    CONSTRAINT email_connectors_validation_status_check CHECK ((validation_status = ANY (ARRAY['pending'::text, 'validating'::text, 'validated'::text, 'failed'::text])))
);


ALTER TABLE public.email_connectors OWNER TO tms;

--
-- Name: email_domain_validations; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_domain_validations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    domain text NOT NULL,
    validation_token text NOT NULL,
    status text DEFAULT 'pending'::text,
    verified_at timestamp with time zone,
    expires_at timestamp with time zone DEFAULT (now() + '24:00:00'::interval) NOT NULL,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT email_domain_validations_status_check CHECK ((status = ANY (ARRAY['pending'::text, 'verified'::text, 'failed'::text, 'expired'::text])))
);


ALTER TABLE public.email_domain_validations OWNER TO tms;

--
-- Name: email_inbox; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_inbox (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid,
    message_id text NOT NULL,
    thread_id text,
    uid integer,
    mailbox_address text NOT NULL,
    from_address text NOT NULL,
    from_name text,
    to_addresses text[] NOT NULL,
    cc_addresses text[],
    bcc_addresses text[],
    reply_to_addresses text[],
    subject text NOT NULL,
    body_text text,
    body_html text,
    snippet text,
    is_read boolean DEFAULT false,
    is_reply boolean DEFAULT false,
    has_attachments boolean DEFAULT false,
    attachment_count integer DEFAULT 0,
    size_bytes integer,
    sent_at timestamp with time zone,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    sync_status text DEFAULT 'synced'::text,
    processing_error text,
    ticket_id uuid,
    is_converted_to_ticket boolean DEFAULT false,
    connector_id uuid,
    headers jsonb,
    raw_email bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT email_inbox_sync_status_check CHECK ((sync_status = ANY (ARRAY['synced'::text, 'processing'::text, 'error'::text])))
);


ALTER TABLE public.email_inbox OWNER TO tms;

--
-- Name: email_mailboxes; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_mailboxes (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    address text NOT NULL,
    inbound_connector_id uuid NOT NULL,
    routing_rules jsonb DEFAULT '[]'::jsonb,
    allow_new_ticket boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    project_id uuid,
    display_name text
);


ALTER TABLE public.email_mailboxes OWNER TO tms;

--
-- Name: COLUMN email_mailboxes.display_name; Type: COMMENT; Schema: public; Owner: tms
--

COMMENT ON COLUMN public.email_mailboxes.display_name IS 'Friendly display name for the mailbox (e.g., "Support Team")';


--
-- Name: email_sync_status; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_sync_status (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    connector_id uuid NOT NULL,
    mailbox_address text NOT NULL,
    last_sync_at timestamp with time zone,
    last_uid integer DEFAULT 0,
    last_message_date timestamp with time zone,
    sync_status text DEFAULT 'idle'::text,
    sync_error text,
    emails_synced_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT email_sync_status_sync_status_check CHECK ((sync_status = ANY (ARRAY['idle'::text, 'syncing'::text, 'error'::text, 'paused'::text])))
);


ALTER TABLE public.email_sync_status OWNER TO tms;

--
-- Name: email_transports; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.email_transports (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    outbound_connector_id uuid NOT NULL,
    envelope_from_domain text,
    dkim_selector text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.email_transports OWNER TO tms;

--
-- Name: file_attachments; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.file_attachments (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid,
    filename character varying(255) NOT NULL,
    original_filename character varying(255) NOT NULL,
    content_type character varying(100) NOT NULL,
    file_size bigint NOT NULL,
    storage_provider public.storage_provider DEFAULT 'minio'::public.storage_provider NOT NULL,
    storage_path text NOT NULL,
    storage_bucket character varying(100) NOT NULL,
    attachment_type public.attachment_type DEFAULT 'ticket_attachment'::public.attachment_type NOT NULL,
    related_entity_type character varying(50),
    related_entity_id uuid,
    checksum character varying(64),
    is_public boolean DEFAULT false NOT NULL,
    expires_at timestamp with time zone,
    uploaded_by_agent_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.file_attachments OWNER TO tms;

--
-- Name: goose_db_version; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.goose_db_version (
    id integer NOT NULL,
    version_id bigint NOT NULL,
    is_applied boolean NOT NULL,
    tstamp timestamp without time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.goose_db_version OWNER TO tms;

--
-- Name: goose_db_version_id_seq; Type: SEQUENCE; Schema: public; Owner: tms
--

ALTER TABLE public.goose_db_version ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.goose_db_version_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: integration_categories; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.integration_categories (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(100) NOT NULL,
    display_name character varying(100) NOT NULL,
    description text,
    icon character varying(100),
    sort_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.integration_categories OWNER TO tms;

--
-- Name: integration_sync_logs; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.integration_sync_logs (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    integration_id uuid NOT NULL,
    operation character varying(100) NOT NULL,
    status character varying(50) NOT NULL,
    external_id character varying(255),
    request_payload jsonb,
    response_payload jsonb,
    error_message text,
    duration_ms integer,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.integration_sync_logs OWNER TO tms;

--
-- Name: integration_templates; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.integration_templates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    category_id uuid NOT NULL,
    type public.integration_type NOT NULL,
    name character varying(255) NOT NULL,
    display_name character varying(255) NOT NULL,
    description text,
    logo_url text,
    website_url text,
    documentation_url text,
    auth_method text NOT NULL,
    config_schema jsonb DEFAULT '{}'::jsonb NOT NULL,
    default_config jsonb DEFAULT '{}'::jsonb NOT NULL,
    supported_events text[] DEFAULT '{}'::text[] NOT NULL,
    is_featured boolean DEFAULT false NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT integration_templates_auth_method_check CHECK ((auth_method = ANY (ARRAY['oauth'::text, 'api_key'::text, 'none'::text])))
);


ALTER TABLE public.integration_templates OWNER TO tms;

--
-- Name: integrations; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.integrations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    type public.integration_type NOT NULL,
    name character varying(255) NOT NULL,
    status public.integration_status DEFAULT 'configuring'::public.integration_status NOT NULL,
    config jsonb DEFAULT '{}'::jsonb NOT NULL,
    oauth_token_id uuid,
    webhook_url text,
    webhook_secret character varying(255),
    last_sync_at timestamp with time zone,
    last_error text,
    retry_count integer DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.integrations OWNER TO tms;

--
-- Name: notification_deliveries; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.notification_deliveries (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    notification_id uuid NOT NULL,
    channel public.notification_channel NOT NULL,
    status character varying(50) DEFAULT 'pending'::character varying NOT NULL,
    external_id character varying(255),
    error_message text,
    delivered_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notification_deliveries OWNER TO tms;

--
-- Name: notifications; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.notifications (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid,
    agent_id uuid NOT NULL,
    type public.notification_type NOT NULL,
    title character varying(255) NOT NULL,
    message text NOT NULL,
    priority public.notification_priority DEFAULT 'normal'::public.notification_priority NOT NULL,
    channels public.notification_channel[] DEFAULT ARRAY['web'::public.notification_channel] NOT NULL,
    action_url text,
    metadata jsonb DEFAULT '{}'::jsonb NOT NULL,
    is_read boolean DEFAULT false NOT NULL,
    read_at timestamp with time zone,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.notifications OWNER TO tms;

--
-- Name: oauth_tokens; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.oauth_tokens (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    provider text NOT NULL,
    account_email text NOT NULL,
    access_token_enc bytea NOT NULL,
    refresh_token_enc bytea,
    expires_at timestamp with time zone NOT NULL,
    scopes text[],
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT oauth_tokens_provider_check CHECK ((provider = ANY (ARRAY['google'::text, 'microsoft'::text])))
);


ALTER TABLE public.oauth_tokens OWNER TO tms;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.organizations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    name character varying(255) NOT NULL,
    external_ref character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.organizations OWNER TO tms;

--
-- Name: projects; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.projects (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    key character varying(50) NOT NULL,
    name character varying(255) NOT NULL,
    status public.project_status DEFAULT 'active'::public.project_status NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.projects OWNER TO tms;

--
-- Name: rate_limit_buckets; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.rate_limit_buckets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    identifier character varying(255) NOT NULL,
    bucket_type character varying(100) NOT NULL,
    current_count integer DEFAULT 0 NOT NULL,
    max_count integer NOT NULL,
    window_start timestamp with time zone NOT NULL,
    window_duration interval NOT NULL,
    last_refill timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.rate_limit_buckets OWNER TO tms;

--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.role_permissions (
    role public.role_type NOT NULL,
    perm character varying(100) NOT NULL
);


ALTER TABLE public.role_permissions OWNER TO tms;

--
-- Name: roles; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.roles (
    role public.role_type NOT NULL,
    description text
);


ALTER TABLE public.roles OWNER TO tms;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.schema_migrations (
    version character varying(255) NOT NULL,
    applied_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE public.schema_migrations OWNER TO tms;

--
-- Name: tenant_project_settings; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.tenant_project_settings (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    setting_key character varying(100) NOT NULL,
    setting_value jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tenant_project_settings OWNER TO tms;

--
-- Name: tenants; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.tenants (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    name character varying(255) NOT NULL,
    status public.tenant_status DEFAULT 'active'::public.tenant_status NOT NULL,
    region character varying(50),
    kms_key_id character varying(255),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tenants OWNER TO tms;

--
-- Name: ticket_mail_routing; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.ticket_mail_routing (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    ticket_id uuid NOT NULL,
    public_token text NOT NULL,
    reply_address text NOT NULL,
    message_id_root text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    revoked_at timestamp with time zone
);


ALTER TABLE public.ticket_mail_routing OWNER TO tms;

--
-- Name: ticket_messages; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.ticket_messages (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    ticket_id uuid NOT NULL,
    author_type public.author_type NOT NULL,
    author_id uuid,
    body text NOT NULL,
    is_private boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.ticket_messages OWNER TO tms;

--
-- Name: ticket_number_seq; Type: SEQUENCE; Schema: public; Owner: tms
--

CREATE SEQUENCE public.ticket_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.ticket_number_seq OWNER TO tms;

--
-- Name: ticket_tags; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.ticket_tags (
    ticket_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    tag character varying(50) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.ticket_tags OWNER TO tms;

--
-- Name: tickets; Type: TABLE; Schema: public; Owner: tms
--

CREATE TABLE public.tickets (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    tenant_id uuid NOT NULL,
    project_id uuid NOT NULL,
    number integer NOT NULL,
    subject character varying(500) NOT NULL,
    status public.ticket_status DEFAULT 'new'::public.ticket_status NOT NULL,
    priority public.ticket_priority DEFAULT 'normal'::public.ticket_priority NOT NULL,
    type public.ticket_type DEFAULT 'question'::public.ticket_type NOT NULL,
    source public.ticket_source DEFAULT 'web'::public.ticket_source NOT NULL,
    customer_id uuid NOT NULL,
    assignee_agent_id uuid,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.tickets OWNER TO tms;

--
-- Data for Name: agent_project_roles; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.agent_project_roles (agent_id, tenant_id, project_id, role, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440031	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	agent	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440032	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	agent	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440033	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	supervisor	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440033	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	supervisor	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440034	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	read_only	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440031	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	agent	2025-08-10 12:23:15.257495+00	2025-08-10 12:23:15.257495+00
0d2d3e71-9114-4302-8bf0-68d25db2e948	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	agent	2025-08-10 18:16:08.944634+00	2025-08-10 18:16:08.944634+00
0d2d3e71-9114-4302-8bf0-68d25db2e948	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	agent	2025-08-10 18:16:29.090115+00	2025-08-10 18:16:29.090115+00
550e8400-e29b-41d4-a716-446655440030	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	tenant_admin	2025-08-10 14:26:00.000576+00	2025-08-21 00:45:28.2047+00
\.


--
-- Data for Name: agents; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.agents (id, tenant_id, email, name, status, password_hash, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440030	550e8400-e29b-41d4-a716-446655440000	admin@acme.com	Admin User	active	$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440031	550e8400-e29b-41d4-a716-446655440000	agent1@acme.com	Alice Agent	active	$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440032	550e8400-e29b-41d4-a716-446655440000	agent2@acme.com	Bob Agent	active	$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440033	550e8400-e29b-41d4-a716-446655440000	supervisor@acme.com	Charlie Supervisor	active	$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440034	550e8400-e29b-41d4-a716-446655440000	readonly@acme.com	David Readonly	active	$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
0d2d3e71-9114-4302-8bf0-68d25db2e948	550e8400-e29b-41d4-a716-446655440000	sumanrocs@gmail.com	Suman Saurabh	active	$2a$10$2CO3unnuisqZ/s0Gd1sBiewW8h0Hl.6b0k3cT1yQraJghe1c/Lpoq	2025-08-10 14:28:33.62762+00	2025-08-10 14:28:33.62762+00
123e4567-e89b-12d3-a456-426614174004	123e4567-e89b-12d3-a456-426614174000	agent@test.com	Support Agent	active	\N	2025-08-15 07:09:23.749913+00	2025-08-15 07:09:23.749913+00
\.


--
-- Data for Name: api_keys; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.api_keys (id, tenant_id, project_id, name, key_hash, key_prefix, scopes, last_used_at, expires_at, is_active, created_by, created_at, updated_at) FROM stdin;
6b1e3b65-3b73-4e39-bdad-cba1568abb51	550e8400-e29b-41d4-a716-446655440000	\N	sdvsdvds	51d8d5dbb9e9e53088a158cbd3cd107cbd51a3b72ede41389249b6dc643366a8	tms_ad8accdd...	{}	\N	\N	t	550e8400-e29b-41d4-a716-446655440030	2025-08-10 10:37:28.215839+00	2025-08-10 10:37:28.215839+00
\.


--
-- Data for Name: attachments; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.attachments (id, tenant_id, project_id, ticket_id, message_id, blob_key, filename, content_type, size_bytes, created_at) FROM stdin;
\.


--
-- Data for Name: chat_messages; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.chat_messages (id, tenant_id, project_id, session_id, message_type, content, author_type, author_id, author_name, metadata, is_private, read_by_visitor, read_by_agent, read_at, created_at) FROM stdin;
1e579156-e28c-4e7e-8012-bf6ae5322e64	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Our agent Admin User has joined the conversation	system	\N	System	{}	f	f	f	\N	2025-08-24 00:15:56.556261+00
9b9b598c-e4f5-42a6-b440-a40b82b88959	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	hello	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 00:16:05.359389+00
9f770da6-f1a7-45e2-a88f-f36646485c47	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hello	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 01:17:19.834715+00
eb518f8d-9c20-45a1-a619-dfc5634e0a14	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hello	visitor	\N	Visitor	{}	f	t	t	2025-08-24 01:18:33.255946+00	2025-08-24 01:18:30.027774+00
a6bc6228-88bf-4dcf-af53-f0f88be0fcef	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	world	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 01:18:42.985636+00
e08e775c-837a-4170-8b52-999ec4341610	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	remo	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 01:18:52.006827+00
8d8d2d90-7da9-48c9-93c7-8138fbe961e1	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	hello	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 02:23:18.057592+00
d2defcc1-7d39-4a44-a18c-f0a5b2b9024f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	0524abd4-b968-4eec-be17-4ca549fdf9c2	text	Hi	visitor	\N	Visitor	{}	f	t	t	2025-08-24 02:25:05.176246+00	2025-08-24 02:24:39.843444+00
eec63cd5-13e9-4857-8198-32c749a1b658	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	0524abd4-b968-4eec-be17-4ca549fdf9c2	text	world	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 02:25:32.941107+00
7a95124a-7bdb-4742-b511-432c38f9e7a6	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hello	visitor	\N	Visitor	{}	f	t	t	2025-08-24 02:29:14.33503+00	2025-08-24 02:29:08.726887+00
afd07172-0b2a-47ae-85ba-f2dfbd3def2d	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hi	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 02:32:09.914759+00
dacefb8f-a1de-4a0d-9e27-99e7476f9c7f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	hel	visitor	\N	Visitor	{}	f	t	t	2025-08-24 02:32:39.447591+00	2025-08-24 02:32:37.58727+00
fab65865-68e8-486b-91aa-e9b3430f363f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	heelllo	visitor	\N	Visitor	{}	f	t	t	2025-08-24 08:55:42.953373+00	2025-08-24 08:55:41.623819+00
6dceb8af-8e67-450c-8eb3-d0ac6919d3cf	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hi	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 09:13:25.456804+00
f34a1935-2321-448d-88ef-86243c26802a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hello	visitor	\N	Visitor	{}	f	t	t	2025-08-24 09:13:35.618048+00	2025-08-24 09:13:33.398167+00
c4cf06c6-fb19-498e-8b85-e85ab54102e3	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Hi	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 09:13:39.787771+00
b3a1027e-0fc0-4443-beda-cf126678c13f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Rambo	agent	550e8400-e29b-41d4-a716-446655440030	Agent	{}	f	f	t	\N	2025-08-24 09:13:45.515789+00
8daf7fb8-68f1-4afd-8687-d9344ff80f24	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	9a3b8671-7078-4e8e-8ec0-b922eb93b625	text	Bomon	visitor	\N	Visitor	{}	f	t	t	2025-08-24 09:13:51.965976+00	2025-08-24 09:13:49.938248+00
\.


--
-- Data for Name: chat_session_participants; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.chat_session_participants (session_id, agent_id, tenant_id, role, joined_at, left_at) FROM stdin;
\.


--
-- Data for Name: chat_sessions; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.chat_sessions (id, tenant_id, project_id, widget_id, customer_id, ticket_id, status, visitor_info, assigned_agent_id, assigned_at, started_at, ended_at, last_activity_at, created_at, updated_at, client_session_id) FROM stdin;
0524abd4-b968-4eec-be17-4ca549fdf9c2	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	f04f21d7-9a80-4f7d-8320-c0d0229c2bff	\N	\N	active	{"language": "en-US", "timezone": "Asia/Calcutta", "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36", "fingerprint": "00cabhjy006r3rbo00kpia7k00ip9bdc0025fel600c0zosb00ipokec00m8t6os", "visitor_name": "Sumania"}	\N	\N	2025-08-24 02:24:35.282647+00	\N	2025-08-24 02:25:32.967523+00	2025-08-24 02:24:35.282647+00	2025-08-24 02:24:35.282647+00	00cabhjy006r3rbo00kpia7k00ip9bdc0025fel600c0zosb00ipokec00m8t6os_1756002275232
9a3b8671-7078-4e8e-8ec0-b922eb93b625	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	f04f21d7-9a80-4f7d-8320-c0d0229c2bff	\N	\N	active	{"language": "en-US", "timezone": "Asia/Calcutta", "user_agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36", "fingerprint": "00uqry8100c11esc009v8swp00oo93yw00uboh6l009xymr000ratf59008nf8no", "visitor_name": "Suman"}	550e8400-e29b-41d4-a716-446655440030	2025-08-24 00:15:56.549384+00	2025-08-24 00:11:03.748015+00	\N	2025-08-24 09:13:49.962852+00	2025-08-24 00:11:03.748015+00	2025-08-24 00:15:56.549384+00	00uqry8100c11esc009v8swp00oo93yw00uboh6l009xymr000ratf59008nf8no_1755992808669
\.


--
-- Data for Name: chat_widgets; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.chat_widgets (id, tenant_id, project_id, domain_id, name, is_active, primary_color, secondary_color, "position", welcome_message, offline_message, auto_open_delay, show_agent_avatars, allow_file_uploads, require_email, business_hours, embed_code, created_at, updated_at, widget_shape, chat_bubble_style, agent_name, agent_avatar_url, use_ai, custom_css, widget_size, animation_style, sound_enabled, show_powered_by, custom_greeting, away_message, background_color, require_name) FROM stdin;
c056e31c-1cfd-409d-92f8-946ca32ae95c	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	4c9d9abd-ebb3-4d00-9142-b33776faec21	Test Chat Widget	t	#2563eb	#f3f4f6	bottom-right	Hello! How can we help you today?	We are currently offline. Please leave a message.	0	t	t	t	{"enabled": false}	\N	2025-08-17 06:05:56.160336+00	2025-08-17 06:05:56.160336+00	rounded	modern	Support Agent	\N	f	\N	medium	smooth	t	t	\N	We're currently away. Leave us a message and we'll get back to you!	#ffffff	f
f04f21d7-9a80-4f7d-8320-c0d0229c2bff	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	657b0834-bbe1-4210-b3de-5e19114dea89	Widhet	t	#d31763	#f3f4f6	bottom-right	Hey!! How can we help you today?	We are currently offline. Please leave a message.	0	t	f	f	{"enabled": false}	<!-- Hith Chat Widget -->\n<script>\n  (function() {\n    window.TMSChatConfig = {\n      widgetId: 'f04f21d7-9a80-4f7d-8320-c0d0229c2bff',\n      domain: 'penify.dev'\n    };\n    var script = document.createElement('script');\n    script.src = 'https://cdn.example.com/chat-widget.js';\n    script.async = true;\n    document.head.appendChild(script);\n  })();\n</script>	2025-08-17 00:48:20.627151+00	2025-08-23 01:13:49.932618+00	rounded	bot	Support Agent	https://media.licdn.com/dms/image/v2/D5603AQEDru6Q4UkzEg/profile-displayphoto-shrink_800_800/profile-displayphoto-shrink_800_800/0/1681498321113?e=1758153600&v=beta&t=3mP9mfYOFW3RuUpMU1qQntWJZPIY3Ie8o2cXq9m5HtY	t	\N	medium	smooth	t	t	Hi there! 👋 How can we help you today?	We're currently away. Leave us a message and we'll get back to you!	#ffffff	t
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.customers (id, tenant_id, email, name, org_id, created_at, updated_at, metadata) FROM stdin;
550e8400-e29b-41d4-a716-446655440020	550e8400-e29b-41d4-a716-446655440000	john.doe@example.com	John Doe	550e8400-e29b-41d4-a716-446655440010	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00	\N
550e8400-e29b-41d4-a716-446655440021	550e8400-e29b-41d4-a716-446655440000	jane.smith@test.com	Jane Smith	550e8400-e29b-41d4-a716-446655440011	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00	\N
550e8400-e29b-41d4-a716-446655440022	550e8400-e29b-41d4-a716-446655440000	bob.wilson@example.com	Bob Wilson	550e8400-e29b-41d4-a716-446655440010	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00	\N
123e4567-e89b-12d3-a456-426614174003	123e4567-e89b-12d3-a456-426614174000	john.doe@example.com	John Doe	\N	2025-08-15 07:09:23.747734+00	2025-08-15 07:09:23.747734+00	\N
313a2882-ac8e-4d91-8458-e991c6ee05c0	550e8400-e29b-41d4-a716-446655440000	Dee - Founder <dee@pearllemongroup.uk>	Dee - Founder	\N	2025-08-15 23:35:13.979681+00	2025-08-15 23:35:13.979681+00	{}
0dd71575-160f-4f80-84a9-4429e826d6c4	550e8400-e29b-41d4-a716-446655440000	Suman Saurabh <sumanrocs@gmail.com>	Suman Saurabh	\N	2025-08-15 23:50:43.519203+00	2025-08-15 23:50:43.519203+00	{}
21ad5048-c647-46fe-8226-37417fedad2c	550e8400-e29b-41d4-a716-446655440000	dee@pearllemongroup.uk	Dee - Founder	\N	2025-08-15 23:56:13.506054+00	2025-08-15 23:56:13.506054+00	{}
900f75ca-de14-4c4b-a5bf-b4b29432f01b	550e8400-e29b-41d4-a716-446655440000	test@example.com	Test User	\N	2025-08-17 06:08:45.42451+00	2025-08-17 06:08:45.42451+00	{}
\.


--
-- Data for Name: email_attachments; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_attachments (id, email_id, tenant_id, filename, content_type, size_bytes, content_id, is_inline, storage_path, storage_url, created_at, project_id) FROM stdin;
\.


--
-- Data for Name: email_connectors; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_connectors (id, tenant_id, type, name, is_active, imap_host, imap_port, imap_use_tls, imap_username, imap_password_enc, imap_folder, imap_seen_strategy, smtp_host, smtp_port, smtp_use_tls, smtp_username, smtp_password_enc, oauth_provider, oauth_account_email, oauth_token_ref, dkim_selector, dkim_public_key, dkim_private_key_enc, return_path_domain, provider_webhook_secret, last_health, created_at, updated_at, project_id, is_validated, validation_status, validation_error, last_validation_at) FROM stdin;
faa8850b-44d4-4258-a45c-abde43abe35f	550e8400-e29b-41d4-a716-446655440000	inbound_imap	Support	t	imap.gmail.com	993	t	sumansaurabh@snorkell.ai	\\x556c477056446b516b484273444e7538786c445a704e496171462b394a664e5a7a5864412f4b4450744d344f7779327431312b3657526176762b7a6b6f52733d	INBOX	mark_seen_after_parse	smtp.gmail.com	587	\N	sumansaurabh@snorkell.ai	\\x76476857746a394e786e593261696f53635549674357626b453436343137573162346a4c6b377a67446c496b66474f594772364d52336337763931493562513d	\N	\N	\N	\N	\N	\\x	\N	\N	{}	2025-08-14 11:29:47.401959+00	2025-08-14 11:37:46.448556+00	550e8400-e29b-41d4-a716-446655440001	t	validated	\N	2025-08-14 11:37:46.448556+00
\.


--
-- Data for Name: email_domain_validations; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_domain_validations (id, tenant_id, project_id, domain, validation_token, status, verified_at, expires_at, metadata, created_at, updated_at) FROM stdin;
657b0834-bbe1-4210-b3de-5e19114dea89	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	penify.dev	34zoqeet0mlu0tlua5dgiouplqxtwj50	verified	2025-08-14 04:51:08.035942+00	2025-08-15 04:50:38.317987+00	{"dns_value": "34zoqeet0mlu0tlua5dgiouplqxtwj50", "dns_record": "_tms-validation.penify.dev"}	2025-08-14 04:50:38.317987+00	2025-08-14 04:51:08.035943+00
85ffb28e-a61b-4d03-91da-7051b5afa922	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	snorkell.ai	jtw379odgzr2yalv0en37kych8ipnblb	verified	2025-08-14 04:52:33.009298+00	2025-08-15 04:51:40.831549+00	{"dns_value": "jtw379odgzr2yalv0en37kych8ipnblb", "dns_record": "_tms-validation.snorkell.ai"}	2025-08-14 04:51:40.831549+00	2025-08-14 04:52:33.009299+00
4c9d9abd-ebb3-4d00-9142-b33776faec21	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	localhost	test-token	verified	\N	2025-08-18 06:05:45.490678+00	{}	2025-08-17 06:05:45.490678+00	2025-08-17 06:05:45.490678+00
\.


--
-- Data for Name: email_inbox; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_inbox (id, tenant_id, project_id, message_id, thread_id, uid, mailbox_address, from_address, from_name, to_addresses, cc_addresses, bcc_addresses, reply_to_addresses, subject, body_text, body_html, snippet, is_read, is_reply, has_attachments, attachment_count, size_bytes, sent_at, received_at, sync_status, processing_error, ticket_id, is_converted_to_ticket, connector_id, headers, raw_email, created_at, updated_at) FROM stdin;
d2738acd-1e2b-4e33-b008-d44206a2c99b	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	2714551518591471170@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 2714551518591471170			\N	f	f	f	0	\N	2024-02-27 23:59:59+00	2025-08-15 23:31:02.832187+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.832187+00	2025-08-15 23:31:02.832187+00
c34050c2-ea3d-473a-9263-d20f47f1aec2	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	d5f76c71-2a09-44b7-b593-42242e8b4331@khadinakbar.com	\N	\N	support@snorkell.ai	Khadin Akbar <me@khadinakbar.com>	Khadin Akbar	{support@snorkell.ai}	\N	\N	\N	Partnership request regarding Udemy course	Hi,\r\n\r\nI’m Khadin Akbar; Udemy Instructor(200k+ Students) and YouTuber(10k+ Subscribers). I create courses and videos on AI tools, side hustles, online businesses, and freelancing, etc.\r\n\r\nI am working on a new course and I want to collaborate with you and want you to sponsor my course, so I will create a dedicated course on your tool.\r\n\r\nMy courses usually receive 2000+ enrollments each month and complete 10k in the first 5 months.\r\n\r\nUdemy: https://udemy.com/user/khadinakbar\r\nYouTube: https://youtube.com/c/growwithkhadin\r\n\r\nLet me know if you're interested.\r\n\r\nI can share my mediakit and more details as well.\r\n\r\nRegards,\r\nKhadin\r\n\r\n\r\nDon't want me to contact you again? Click here https://inst.khadinakbar.com/unsub/1/aa019e9e-bdd3-4997-bd66-a0788741036c	<div>Hi,</div><div><br>I&rsquo;m Khadin Akbar; Udemy Instructor(200k+ Students) and YouTuber(10k+ Subscribers). I create courses and videos on AI tools, side hustles, online businesses, and freelancing, etc.</div><div><br>I am working on a new course and I want to collaborate with you and want you to sponsor my course, so I will create a dedicated course on your tool.</div><div><br>My courses usually receive 2000+ enrollments each month and complete 10k in the first 5 months.</div><div><br></div><div>Udemy: <a target="_blank" rel="noopener noreferrer" href="https://udemy.com/user/khadinakbar">https://udemy.com/user/khadinakbar</a></div><div>YouTube: <a target="_blank" rel="noopener noreferrer noopener noreferrer" href="https://youtube.com/c/growwithkhadin">https://youtube.com/c/growwithkhadin</a></div><div><br>Let me know if you're interested.</div><div><br></div><div>I can share my mediakit and more details as well.</div><div><br>Regards,</div><div>Khadin</div><div><br></div><div><br></div><div><span style="font-size: 12px;">Don't want me to contact you again?&nbsp;</span><a href="https://inst.khadinakbar.com/unsub/1/aa019e9e-bdd3-4997-bd66-a0788741036c" target="_blank" rel="noopener noreferrer"><span style="font-size: 12px;">Click here</span></a></div>	Hi, I’m Khadin Akbar; Udemy Instructor(200k+ Students) and YouTuber(10k+ Subscribers). I create courses and videos on AI tools, side hustles, online...	f	f	f	0	\N	2024-06-07 15:14:54+00	2025-08-15 23:31:02.839683+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.839683+00	2025-08-15 23:31:02.839683+00
e6193f78-de28-43a9-a374-3f3204160a7f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAGbEnoyn9EtvNc_DA5YzsE0yuqYqBzpO8YMCUzhUBcNSmAKoEQ@mail.gmail.com	\N	\N	support@snorkell.ai	WeSubmitYourStartup <wesubmityourstartup5@gmail.com>	WeSubmitYourStartup	{support@snorkell.ai}	\N	\N	\N	We manually submit your startup to all Free & Freminum Startup Directories - 179 Directories	Hello,\r\n\r\nI saw that you're working on a startup. I work as a freelancer, where I\r\nhelp people like you achieve their goals with the help of *Startup\r\nSubmission in Startup Directories*. If you have just finished building your\r\ncompany or product website, you probably want to do some initial promotion\r\nor online visibility. Hence, I will submit your product website to 100+\r\nstartup directories to get the ball rolling.\r\n\r\nThey are the best places to post your product/startup/app and get feedback.\r\n\r\n\r\n\r\nMany platforms and websites are available to help startups engage with\r\ntheir target users and gain further product validation and reach\r\nproduct/market fit quickly.\r\n\r\n\r\n\r\nContent to sign up to startup directories :\r\n\r\n\r\n\r\n  1.   Startup name and website and Email id to use while submission\r\n  2.   Your logo and Screenshots\r\n  3.   Your location\r\n  4.   Founder's Name and Brief about Founder's\r\n  5.   Logo and Screenshots Url\r\n  6.   Brief description (6 Words)\r\n  7.   Long description (4 Sentence)\r\n  8.   Problem your startup solves\r\n  9.   Target Audience\r\n 10.  Categories belongs to your startup\r\n 11.  Number of Employees\r\n 12.  Social media handles\r\n 13.  Found/Launch date\r\n\r\nBenefits :\r\n\r\n*Gain Digital Exposure* - Improve the online presence and reputation of\r\nyour brand locally and globally.\r\n*Earn Quality Backlinks* - Backlinks represent a vote of confidence for\r\ndifferent search engines. With SR Booster, you will get quality backlinks\r\nfrom authority domains.\r\n*Attract Investors & Clients* - Listing your startup on directories is an\r\nideal place to find business partners and connect to qualified leads.\r\n*Get Feedback* - Some directories & blogs from our list, allow users and\r\nyour customers to give valuable feedback for your project/company.\r\n*Save Time* - Save more than 30 hours of boring manual submissions.\r\n\r\nI know you're busy, so here are a few quick features that i think you might\r\nfind useful :\r\n\r\n1. We will submit your startup to 179 different startup directories.\r\n2. Your startup is listed on them.\r\n3. Interested people or customers know about your startup.\r\n4. Give your startup good traffic.\r\n5. Improve your SEO & Page rank.\r\n\r\nI would like to let you know that the submissions are done as promised and\r\nthe report will be sent to your submitted mail i'd.\r\n\r\nHow it works :\r\n\r\nWe will take details of your startup from you and manually submit it to\r\nthese websites. Also, Some of the websites require an account for\r\nsubmitting to them. In this case, we will create an account for you and you\r\nwill receive all username/password generated by us in the submission report\r\n\r\nNote :-  Report will be shared within 48hrs.\r\n\r\nPricing : $99 after the submission in 179 startup directories and once the\r\nreport shared with screenshot proof.\r\n\r\nAbout us : As a freelancer I am trying to achieve my goal of success with\r\nthis project (startup submission in startup directories). One day I will\r\nalso be a startup founder :)\r\n\r\nLooking forward & eager to get started on your project.\r\n\r\nThank you,\r\n	<div dir="ltr"><font face="arial, sans-serif">Hello,</font><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I saw that you&#39;re working on a startup. I work as a freelancer, where I help people like you achieve their goals with the help of <b>Startup Submission in Startup Directories</b>. <font color="#000000">If you have just finished building your company or product website, you probably want to do some initial promotion or online visibility. Hence, I will submit your product website to 100+ startup directories to get the ball rolling.</font><span style="color:rgb(98,100,106)"> </span></font></div><div><span style="color:rgb(98,100,106)"><font face="arial, sans-serif"><br></font></span></div><div><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">They are the best places to post your product/startup/app and get feedback.</font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"> </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">Many platforms and websites are available to help startups engage with their target users and gain further product validation and reach product/market fit quickly. </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"> </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">Content to sign up to startup directories :</font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"><br></font></p><p class="MsoNormal" style="margin:0in 3.45pt 0.0001pt 19.9pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">  </font></p><span style="color:rgb(0,0,0)">  1.   Startup name and website and Email id to use while submission</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  2.   Your logo and Screenshots</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  3.   Your location</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  4.   Founder&#39;s Name and Brief about Founder&#39;s</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  5.   Logo and Screenshots Url</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  6.   Brief description (6 Words)</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  7.   Long description (4 Sentence)</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  8.   Problem your startup solves</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  9.   Target Audience</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 10.  Categories belongs to your startup</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 11.  Number of Employees  </span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 12.  Social media handles</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 13.  Found/Launch date</span></div><div><span style="color:rgb(0,0,0)"><br></span></div><div><span style="color:rgb(0,0,0)">Benefits :</span></div><div><span style="color:rgb(0,0,0)"><br></span></div><div><b>Gain Digital Exposure</b> - Improve the online presence and reputation of your brand locally and globally.<br><b>Earn Quality Backlinks</b> - Backlinks represent a vote of confidence for different search engines. With SR Booster, you will get quality backlinks from authority domains.<br><b>Attract Investors &amp; Clients</b> - Listing your startup on directories is an ideal place to find business partners and connect to qualified leads.<br><b>Get Feedback</b> - Some directories &amp; blogs from our list, allow users and your customers to give valuable feedback for your project/company.<br><b>Save Time</b> - Save more than 30 hours of boring manual submissions.<span style="color:rgb(0,0,0)"><br></span></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I know you&#39;re busy, so here are a few quick features that i think you might find useful :</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">1. We will submit your startup to 179 different startup directories.</font></div><div><span style="font-family:arial,sans-serif">2. Your startup is listed on them.</span><br></div><div><span style="font-family:arial,sans-serif">3. Interested people or customers know about your startup.</span><br></div><div><span style="font-family:arial,sans-serif">4. Give your startup good traffic.</span><br></div><div><span style="font-family:arial,sans-serif">5. Improve your SEO &amp; Page rank.</span><br></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I would like to let you know that the submissions are done as promised and the report will be sent to your submitted mail i&#39;d.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">How it works :</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">We will take details of your startup from you and manually submit it to these websites. Also, Some of the websites require an account for submitting to them. In this case, we will create an account for you and you will receive all username/password generated by us in the submission report</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Note :-  Report will be shared within 48hrs.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Pricing : $99 after the submission in 179 startup directories and once the report shared with screenshot proof.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">About us : As a freelancer I am trying to achieve my goal of success with this project (startup submission in startup directories). One day I will also be a startup founder :)</font></div><div><font face="arial, sans-serif"><br></font></div><div><span style="font-family:arial,sans-serif">Looking forward &amp; eager to get started on your project.</span><br></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Thank you,</font></div></div>\r\n	Hello, I saw that you're working on a startup. I work as a freelancer, where I help people like you achieve their goals with the help of *Startup Subm...	f	f	f	0	\N	2024-04-15 13:27:00+00	2025-08-15 23:31:02.872167+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.872167+00	2025-08-15 23:31:02.872167+00
ddc47968-f9c2-4c16-9b5f-d1634765e5a0	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAGbEnowO-MZH=5Ez_uTaadvHZdjJDh4UsoK0KBoZHRXoAPL_Kg@mail.gmail.com	<CAO+uLtUuaqGFQPYBZ-L-cUsE1P_xM0GstjDBeBCRjvVUNK8OJw@mail.gmail.com>	\N	support@snorkell.ai	WeSubmitYourStartup <wesubmityourstartup5@gmail.com>	WeSubmitYourStartup	{"Suman Saurabh <sumansaurabh@snorkell.ai>"}	{support@snorkell.ai}	\N	\N	Re: We manually submit your startup to all Free & Freminum Startup Directories - 179 Directories	Hello,\r\n\r\nThanks for reaching back.\r\n\r\nCan you send the invitation ?\r\n\r\nAnd also I am attaching the directory list and recently done startup\r\ndetails with their response after the submission.\r\n\r\nNote :- We will skip some to make it 179 directories :\r\n\r\nOnce again thanks for the opportunity.\r\n\r\nLooking forward to hearing from you to proceed with the submission.\r\n\r\nThanks,\r\n\r\nThanks,\r\n\r\nOn Mon, Apr 15, 2024 at 7:05 PM Suman Saurabh <sumansaurabh@snorkell.ai>\r\nwrote:\r\n\r\n> I think that we will be great, we can connect. Can you please schedule an\r\n> invite. Thanks!\r\n>\r\n> On Mon, 15 Apr 2024 at 6:58 PM, WeSubmitYourStartup <\r\n> wesubmityourstartup5@gmail.com> wrote:\r\n>\r\n>> Hello,\r\n>>\r\n>> I saw that you're working on a startup. I work as a freelancer, where I\r\n>> help people like you achieve their goals with the help of *Startup\r\n>> Submission in Startup Directories*. If you have just finished building\r\n>> your company or product website, you probably want to do some initial\r\n>> promotion or online visibility. Hence, I will submit your product website\r\n>> to 100+ startup directories to get the ball rolling.\r\n>>\r\n>> They are the best places to post your product/startup/app and get\r\n>> feedback.\r\n>>\r\n>>\r\n>>\r\n>> Many platforms and websites are available to help startups engage with\r\n>> their target users and gain further product validation and reach\r\n>> product/market fit quickly.\r\n>>\r\n>>\r\n>>\r\n>> Content to sign up to startup directories :\r\n>>\r\n>>\r\n>>\r\n>>   1.   Startup name and website and Email id to use while submission\r\n>>   2.   Your logo and Screenshots\r\n>>   3.   Your location\r\n>>   4.   Founder's Name and Brief about Founder's\r\n>>   5.   Logo and Screenshots Url\r\n>>   6.   Brief description (6 Words)\r\n>>   7.   Long description (4 Sentence)\r\n>>   8.   Problem your startup solves\r\n>>   9.   Target Audience\r\n>>  10.  Categories belongs to your startup\r\n>>  11.  Number of Employees\r\n>>  12.  Social media handles\r\n>>  13.  Found/Launch date\r\n>>\r\n>> Benefits :\r\n>>\r\n>> *Gain Digital Exposure* - Improve the online presence and reputation of\r\n>> your brand locally and globally.\r\n>> *Earn Quality Backlinks* - Backlinks represent a vote of confidence for\r\n>> different search engines. With SR Booster, you will get quality backlinks\r\n>> from authority domains.\r\n>> *Attract Investors & Clients* - Listing your startup on directories is\r\n>> an ideal place to find business partners and connect to qualified leads.\r\n>> *Get Feedback* - Some directories & blogs from our list, allow users and\r\n>> your customers to give valuable feedback for your project/company.\r\n>> *Save Time* - Save more than 30 hours of boring manual submissions.\r\n>>\r\n>> I know you're busy, so here are a few quick features that i think you\r\n>> might find useful :\r\n>>\r\n>> 1. We will submit your startup to 179 different startup directories.\r\n>> 2. Your startup is listed on them.\r\n>> 3. Interested people or customers know about your startup.\r\n>> 4. Give your startup good traffic.\r\n>> 5. Improve your SEO & Page rank.\r\n>>\r\n>> I would like to let you know that the submissions are done as\r\n>> promised and the report will be sent to your submitted mail i'd.\r\n>>\r\n>> How it works :\r\n>>\r\n>> We will take details of your startup from you and manually submit it to\r\n>> these websites. Also, Some of the websites require an account for\r\n>> submitting to them. In this case, we will create an account for you and you\r\n>> will receive all username/password generated by us in the submission report\r\n>>\r\n>> Note :-  Report will be shared within 48hrs.\r\n>>\r\n>> Pricing : $99 after the submission in 179 startup directories and once\r\n>> the report shared with screenshot proof.\r\n>>\r\n>> About us : As a freelancer I am trying to achieve my goal of success with\r\n>> this project (startup submission in startup directories). One day I will\r\n>> also be a startup founder :)\r\n>>\r\n>> Looking forward & eager to get started on your project.\r\n>>\r\n>> Thank you,\r\n>>\r\n>\r\n	<div dir="ltr">Hello,<div><br>Thanks for reaching back.</div><div><br></div><div>Can you send the invitation ?<br><div><br></div><div><div><div>And also I am attaching the directory list and recently done startup details with their response after the submission.<br></div><div><br></div><div>Note :- We will skip some to make it 179 directories :</div><div><br></div><div>Once again thanks for the opportunity.<br></div><div><br>Looking forward to hearing from you to proceed with the submission.</div><div><br></div><div>Thanks,</div></div></div></div><div><br></div><div>Thanks,</div></div><br><div class="gmail_quote"><div dir="ltr" class="gmail_attr">On Mon, Apr 15, 2024 at 7:05 PM Suman Saurabh &lt;<a href="mailto:sumansaurabh@snorkell.ai">sumansaurabh@snorkell.ai</a>&gt; wrote:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex"><div dir="auto">I think that we will be great, we can connect. Can you please schedule an invite. Thanks!</div><div><br><div class="gmail_quote"><div dir="ltr" class="gmail_attr">On Mon, 15 Apr 2024 at 6:58 PM, WeSubmitYourStartup &lt;<a href="mailto:wesubmityourstartup5@gmail.com" target="_blank">wesubmityourstartup5@gmail.com</a>&gt; wrote:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex"><div dir="ltr"><font face="arial, sans-serif">Hello,</font><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I saw that you&#39;re working on a startup. I work as a freelancer, where I help people like you achieve their goals with the help of <b>Startup Submission in Startup Directories</b>. <font color="#000000">If you have just finished building your company or product website, you probably want to do some initial promotion or online visibility. Hence, I will submit your product website to 100+ startup directories to get the ball rolling.</font><span style="color:rgb(98,100,106)"> </span></font></div><div><span style="color:rgb(98,100,106)"><font face="arial, sans-serif"><br></font></span></div><div><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">They are the best places to post your product/startup/app and get feedback.</font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"> </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">Many platforms and websites are available to help startups engage with their target users and gain further product validation and reach product/market fit quickly. </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"> </font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">Content to sign up to startup directories :</font></p><p class="MsoNormal" style="margin:0in 0in 0.0001pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif"><br></font></p><p class="MsoNormal" style="margin:0in 3.45pt 0.0001pt 19.9pt;line-height:normal;background-image:initial;background-position:initial;background-size:initial;background-repeat:initial;background-origin:initial;background-clip:initial"><font color="#000000" face="arial, sans-serif">  </font></p><span style="color:rgb(0,0,0)">  1.   Startup name and website and Email id to use while submission</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  2.   Your logo and Screenshots</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  3.   Your location</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  4.   Founder&#39;s Name and Brief about Founder&#39;s</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  5.   Logo and Screenshots Url</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  6.   Brief description (6 Words)</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  7.   Long description (4 Sentence)</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  8.   Problem your startup solves</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)">  9.   Target Audience</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 10.  Categories belongs to your startup</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 11.  Number of Employees  </span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 12.  Social media handles</span><br style="color:rgb(0,0,0)"><span style="color:rgb(0,0,0)"> 13.  Found/Launch date</span></div><div><span style="color:rgb(0,0,0)"><br></span></div><div><span style="color:rgb(0,0,0)">Benefits :</span></div><div><span style="color:rgb(0,0,0)"><br></span></div><div><b>Gain Digital Exposure</b> - Improve the online presence and reputation of your brand locally and globally.<br><b>Earn Quality Backlinks</b> - Backlinks represent a vote of confidence for different search engines. With SR Booster, you will get quality backlinks from authority domains.<br><b>Attract Investors &amp; Clients</b> - Listing your startup on directories is an ideal place to find business partners and connect to qualified leads.<br><b>Get Feedback</b> - Some directories &amp; blogs from our list, allow users and your customers to give valuable feedback for your project/company.<br><b>Save Time</b> - Save more than 30 hours of boring manual submissions.<span style="color:rgb(0,0,0)"><br></span></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I know you&#39;re busy, so here are a few quick features that i think you might find useful :</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">1. We will submit your startup to 179 different startup directories.</font></div><div><span style="font-family:arial,sans-serif">2. Your startup is listed on them.</span><br></div><div><span style="font-family:arial,sans-serif">3. Interested people or customers know about your startup.</span><br></div><div><span style="font-family:arial,sans-serif">4. Give your startup good traffic.</span><br></div><div><span style="font-family:arial,sans-serif">5. Improve your SEO &amp; Page rank.</span><br></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">I would like to let you know that the submissions are done as promised and the report will be sent to your submitted mail i&#39;d.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">How it works :</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">We will take details of your startup from you and manually submit it to these websites. Also, Some of the websites require an account for submitting to them. In this case, we will create an account for you and you will receive all username/password generated by us in the submission report</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Note :-  Report will be shared within 48hrs.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Pricing : $99 after the submission in 179 startup directories and once the report shared with screenshot proof.</font></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">About us : As a freelancer I am trying to achieve my goal of success with this project (startup submission in startup directories). One day I will also be a startup founder :)</font></div><div><font face="arial, sans-serif"><br></font></div><div><span style="font-family:arial,sans-serif">Looking forward &amp; eager to get started on your project.</span><br></div><div><font face="arial, sans-serif"><br></font></div><div><font face="arial, sans-serif">Thank you,</font></div></div>\r\n</blockquote></div></div>\r\n</blockquote></div>\r\n	Hello, Thanks for reaching back. Can you send the invitation ? And also I am attaching the directory list and recently done startup details with their...	f	t	f	0	\N	2024-04-17 01:53:22+00	2025-08-15 23:31:02.878563+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.878563+00	2025-08-15 23:31:02.878563+00
eb1f43e9-a136-49ee-8b3d-9f86d5f3709c	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	vUErtRgfQheHMIIHHAV02A@geopod-ismtpd-3	\N	\N	support@snorkell.ai	Ana from Avian.io <info@avian.io>	Ana from Avian.io	{support@snorkell.ai}	\N	\N	\N	New Llama 3.3 70B beats GPT 4o and is 3x cheaper, available now on Avian.io API		<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">\r\n<html>\r\n  <head>\r\n    <meta http-equiv="x-ua-compatible" content="ie=edge" />\r\n    <meta name="x-apple-disable-message-reformatting" />\r\n    <meta name="viewport" content="width=device-width, initial-scale=1" />\r\n    <meta\r\n      name="format-detection"\r\n      content="telephone=no, date=no, address=no, email=no"\r\n    />\r\n    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />\r\n    <style type="text/css">\r\n      body,\r\n      table,\r\n      td {\r\n        font-family: Helvetica, Arial, sans-serif !important;\r\n      }\r\n      .ExternalClass {\r\n        width: 100%;\r\n      }\r\n      .ExternalClass,\r\n      .ExternalClass p,\r\n      .ExternalClass span,\r\n      .ExternalClass font,\r\n      .ExternalClass td,\r\n      .ExternalClass div {\r\n        line-height: 150%;\r\n      }\r\n      a {\r\n        text-decoration: none;\r\n      }\r\n      * {\r\n        color: inherit;\r\n      }\r\n      a[x-apple-data-detectors],\r\n      u + #body a,\r\n      #MessageViewBody a {\r\n        color: inherit;\r\n        text-decoration: none;\r\n        font-size: inherit;\r\n        font-family: inherit;\r\n        font-weight: inherit;\r\n        line-height: inherit;\r\n      }\r\n      img {\r\n        -ms-interpolation-mode: bicubic;\r\n      }\r\n      table:not([class^="s-"]) {\r\n        font-family: Helvetica, Arial, sans-serif;\r\n        mso-table-lspace: 0pt;\r\n        mso-table-rspace: 0pt;\r\n        border-spacing: 0px;\r\n        border-collapse: collapse;\r\n      }\r\n      table:not([class^="s-"]) td {\r\n        border-spacing: 0px;\r\n        border-collapse: collapse;\r\n      }\r\n      @media screen and (max-width: 600px) {\r\n        .w-full,\r\n        .w-full > tbody > tr > td {\r\n          width: 100% !important;\r\n        }\r\n        .w-12,\r\n        .w-12 > tbody > tr > td {\r\n          width: 48px !important;\r\n        }\r\n        .w-48,\r\n        .w-48 > tbody > tr > td {\r\n          width: 192px !important;\r\n        }\r\n        .pt-5:not(table),\r\n        .pt-5:not(.btn) > tbody > tr > td,\r\n        .pt-5.btn td a,\r\n        .py-5:not(table),\r\n        .py-5:not(.btn) > tbody > tr > td,\r\n        .py-5.btn td a {\r\n          padding-top: 20px !important;\r\n        }\r\n        .pb-5:not(table),\r\n        .pb-5:not(.btn) > tbody > tr > td,\r\n        .pb-5.btn td a,\r\n        .py-5:not(table),\r\n        .py-5:not(.btn) > tbody > tr > td,\r\n        .py-5.btn td a {\r\n          padding-bottom: 20px !important;\r\n        }\r\n        *[class*="s-lg-"] > tbody > tr > td {\r\n          font-size: 0 !important;\r\n          line-height: 0 !important;\r\n          height: 0 !important;\r\n        }\r\n        .s-2 > tbody > tr > td {\r\n          font-size: 8px !important;\r\n          line-height: 8px !important;\r\n          height: 8px !important;\r\n        }\r\n        .s-6 > tbody > tr > td {\r\n          font-size: 24px !important;\r\n          line-height: 24px !important;\r\n          height: 24px !important;\r\n        }\r\n        .s-8 > tbody > tr > td {\r\n          font-size: 32px !important;\r\n          line-height: 32px !important;\r\n          height: 32px !important;\r\n        }\r\n        .s-10 > tbody > tr > td {\r\n          font-size: 40px !important;\r\n          line-height: 40px !important;\r\n          height: 40px !important;\r\n        }\r\n      }\r\n    </style>\r\n  </head>\r\n  <body\r\n    class="bg-light"\r\n    style="\r\n      outline: 0;\r\n      width: 100%;\r\n      min-width: 100%;\r\n      height: 100%;\r\n      -webkit-text-size-adjust: 100%;\r\n      -ms-text-size-adjust: 100%;\r\n      font-family: Helvetica, Arial, sans-serif;\r\n      line-height: 24px;\r\n      font-weight: normal;\r\n      font-size: 16px;\r\n      -moz-box-sizing: border-box;\r\n      -webkit-box-sizing: border-box;\r\n      box-sizing: border-box;\r\n      color: #000000;\r\n      margin: 0;\r\n      padding: 0;\r\n      border-width: 0;\r\n    "\r\n    bgcolor="#f7fafc"\r\n  >\r\n    <table\r\n      class="bg-light body"\r\n      valign="top"\r\n      role="presentation"\r\n      border="0"\r\n      cellpadding="0"\r\n      cellspacing="0"\r\n      style="\r\n        outline: 0;\r\n        width: 100%;\r\n        min-width: 100%;\r\n        height: 100%;\r\n        -webkit-text-size-adjust: 100%;\r\n        -ms-text-size-adjust: 100%;\r\n        font-family: Helvetica, Arial, sans-serif;\r\n        line-height: 24px;\r\n        font-weight: normal;\r\n        font-size: 16px;\r\n        -moz-box-sizing: border-box;\r\n        -webkit-box-sizing: border-box;\r\n        box-sizing: border-box;\r\n        color: #000000;\r\n        margin: 0;\r\n        padding: 0;\r\n        border-width: 0;\r\n      "\r\n      bgcolor="#f7fafc"\r\n    >\r\n      <tbody>\r\n        <tr>\r\n          <td\r\n            valign="top"\r\n            style="line-height: 24px; font-size: 16px; margin: 0"\r\n            align="left"\r\n            bgcolor="#f7fafc"\r\n          >\r\n            <table\r\n              class="s-8 w-full"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n              width="100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    style="\r\n                      line-height: 32px;\r\n                      font-size: 32px;\r\n                      width: 100%;\r\n                      height: 32px;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                    width="100%"\r\n                    height="32"\r\n                  >\r\n                    &#160;\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n            <table\r\n              class="container"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    align="center"\r\n                    style="\r\n                      line-height: 24px;\r\n                      font-size: 16px;\r\n                      margin: 0;\r\n                      padding: 0 16px;\r\n                    "\r\n                  >\r\n                    <!--[if (gte mso 9)|(IE)]>\r\n                      <table align="center" role="presentation">\r\n                        <tbody>\r\n                          <tr>\r\n                            <td width="600">\r\n                    <![endif]-->\r\n                    <table\r\n                      align="center"\r\n                      role="presentation"\r\n                      border="0"\r\n                      cellpadding="0"\r\n                      cellspacing="0"\r\n                      style="width: 100%; max-width: 600px; margin: 0 auto"\r\n                    >\r\n                      <tbody>\r\n                        <tr>\r\n                          <td\r\n                            style="\r\n                              line-height: 24px;\r\n                              font-size: 16px;\r\n                              margin: 0;\r\n                            "\r\n                            align="left"\r\n                          >\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div class="">\r\n                              <table\r\n                                class="ax-center"\r\n                                role="presentation"\r\n                                align="center"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="margin: 0 auto"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 24px;\r\n                                        font-size: 16px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                    >\r\n                                      <img\r\n                                        class="w-24"\r\n                                        src="https://avian.io/static/images/logomark.png"\r\n                                        alt="Avian Logo"\r\n                                        style="\r\n                                          height: auto;\r\n                                          line-height: 100%;\r\n                                          outline: none;\r\n                                          text-decoration: none;\r\n                                          display: block;\r\n                                          width: 48px;\r\n                                          border-style: none;\r\n                                          border-width: 0;\r\n                                        "\r\n                                        width="48"\r\n                                      />\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <table\r\n                              class="s-10 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 40px;\r\n                                      font-size: 40px;\r\n                                      width: 100%;\r\n                                      height: 40px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="40"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n\r\n                            \r\n\r\n<table\r\n  class="card"\r\n  role="presentation"\r\n  border="0"\r\n  cellpadding="0"\r\n  cellspacing="0"\r\n  style="\r\n    border-radius: 6px;\r\n    border-collapse: separate !important;\r\n    width: 100%;\r\n    overflow: hidden;\r\n    border: 1px solid #e2e8f0;\r\n  "\r\n  bgcolor="#ffffff"\r\n>\r\n  <tbody>\r\n    <tr>\r\n      <td\r\n        style="line-height: 24px; font-size: 16px; width: 100%; margin: 0"\r\n        align="left"\r\n        bgcolor="#ffffff"\r\n      >\r\n        <table\r\n          class="card-body"\r\n          role="presentation"\r\n          border="0"\r\n          cellpadding="0"\r\n          cellspacing="0"\r\n          style="width: 100%"\r\n        >\r\n          <tbody>\r\n            <tr>\r\n              <td\r\n                style="\r\n                  line-height: 24px;\r\n                  font-size: 16px;\r\n                  width: 100%;\r\n                  margin: 0;\r\n                  padding: 40px;\r\n                "\r\n                align="left"\r\n              >\r\n                <h1\r\n                  class="h3"\r\n                  style="\r\n                    padding-top: 0;\r\n                    padding-bottom: 0;\r\n                    font-weight: 500;\r\n                    vertical-align: baseline;\r\n                    font-size: 28px;\r\n                    line-height: 33.6px;\r\n                    margin: 0;\r\n                  "\r\n                  align="left"\r\n                >\r\n              Use the Worlds Fastest Llama 3.3 70B Instruct on at <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3Dg8pr_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kYngRmSIlWkAGEUnpZbwe9pugIjoaOQsL00aJMECm8SAY-2FO4VbMxBMzg5TZOwOJTPzlv-2FAwreGGRlECvfytWkID6RjfjtAvDncbjps9JMok-2FdvccteWFVByx-2Fqac8CyDVXPC7Bkru5dK-2FuMEsL2HD0sv1C052mElOSsOuyznEcPfg-3D-3D">new.avian.io </a> </h1>\r\n                <table\r\n                  class="s-2 w-full"\r\n                  role="presentation"\r\n                  border="0"\r\n                  cellpadding="0"\r\n                  cellspacing="0"\r\n                  style="width: 100%"\r\n                  width="100%"\r\n                >\r\n                  <tbody>\r\n                    <tr>\r\n                      <td\r\n                        style="\r\n                          line-height: 8px;\r\n                          font-size: 8px;\r\n                          width: 100%;\r\n                          height: 8px;\r\n                          margin: 0;\r\n                        "\r\n                        align="left"\r\n                        width="100%"\r\n                        height="8"\r\n                      >\r\n                        &#160;\r\n                      </td>\r\n                    </tr>\r\n                  </tbody>\r\n                </table>\r\n                <table\r\n                  class="py-2"\r\n                  role="presentation"\r\n                  border="0"\r\n                  cellpadding="0"\r\n                  cellspacing="0"\r\n                >\r\n                  <tbody>\r\n                    <tr>\r\n                      <td\r\n                        style="\r\n                          line-height: 24px;\r\n                          font-size: 16px;\r\n                          padding-bottom: 20px;\r\n                          margin: 0;\r\n                        "\r\n                        align="left"\r\n                      >\r\n                        <div class="">\r\n                          <p\r\n                            class="text-gray-700"\r\n                            style="\r\n                              line-height: 24px;\r\n                              font-size: 20px;\r\n                              color: #4a5568;\r\n                              width: 100%;\r\n                              margin: 0;\r\n                            "\r\n                            align="left"\r\n                          ></p>\r\n                          <div><a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3DTcwh_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kYngRmSIlWkAGEUnpZbwe9pwKU5vJrpPh1VKnDZR5fSOsZ7uRMxRM2e3hsPxMj6YlD561JpiYXKaEjQ73uWwHSpcqdZ8GrwHgWpTQhovUN-2FdZIaV3xoS41ic8GKq66fHRTnAz-2FGLPJ-2FC8PgYqoJ-2Fc-2FW9IH4sC4tSSMQXc9RM3jS5g-3D-3D"><img src="https://i.imgur.com/Pu08VM1.png" alt="Avian Llama 3.3 70B record"></a> <div> <p><strong>Llama 3.3 released by Meta has achieved better performance than GPT 4o on many benchmarks, and is available at up to three times the speed and a third of the price. </strong></p> <p> You can use the fastest Llama 3.3 70B at scale on the <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3D7Loq_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kYngRmSIlWkAGEUnpZbwe9pP8RoNDPvrEcTC47j3xP8O0xltbaoxVfZ-2F-2Bs-2BTrrudL-2BHP4LVxgxZ96UyGIsZWCV-2BPhCZaURKSBqxBFzwRgNt-2B7SdKkZ312GI13SPe2MvbvCcrKsZ9Nn0dLTZso-2FVdIuDHBRS6VIMVN2HtjPj0HkYbA-3D-3D">Avian.io API.</a></p> <p> Compared to other providers, we offer superior text generation quality, higher throughput and lower time to first token.</p><p> We've decided to launch at a competitive price of $0.9 per million tokens, and the model is integrated into all of Avian's system as of now.</p> <p> Unlike services like Groq and Cerebras, we're launching this service with no rate limits, so you can hit the ground running \r\n                            and scale to billions of requests. </p>\r\n\r\n\r\n                          <p>We also provide volume discounts for clients doing greater than 1 billion tokens per day. Please get in touch with any questions, or queries about our bespoke finetuning and inference stack than can speed up your LLMs by 5-20x: <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DMP0utklEKqkkbx1xPVUfEPeeFaNeotwZeJr2pawexzOLEp1_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kYngRmSIlWkAGEUnpZbwe9pZY5ea-2FOCWTwhryEmdpstBaw-2FuB5biU-2BloXl6R8i8sWjvF6cOSvbOl5mE1n122wdzSrDMqnuOHgnxfBpdoBywGnnm4P-2FkoUercsSzu-2F3XHPb4LuSRib-2B2m8TylxFZDAsz0cPROSwUR6doG3XWgXRiQQ-3D-3D" style="color: blue;"> Book a Chat With Us</a> </p>\r\n                            \r\n                            The Avian Team</p> </div>\r\n                      </td>\r\n                    </tr>\r\n                  </tbody>\r\n                </table>\r\n                <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3DunK6_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kYngRmSIlWkAGEUnpZbwe9p7GChsu-2F-2BIpiWgZ-2FV9pQQJ9nvZv2JzfDrlmf4c56vUh5QSNCA4yaZV8zRwTG36Vw8iIXsfZvGFUHd5QVX7un-2FpDj78PNlAg5Q-2BCy-2BGG-2FJzYtLTH3pPPCYFZpL7DSLI9T4zbYWCQBLVVII1WiCUIlsDQ-3D-3D">\r\n                  <table\r\n                    class="btn btn-primary"\r\n                    role="presentation"\r\n                    border="0"\r\n                    cellpadding="0"\r\n                    cellspacing="0"\r\n                    style="\r\n                      border-radius: 6px;\r\n                      border-collapse: separate!important;\r\n                      background-color: #007bff; /* blue background color */\r\n                      padding: 10px 20px; /* add some padding */\r\n                      text-align: center; /* center the text */\r\n                      cursor: pointer; /* change cursor to pointer on hover */\r\n                    "\r\n                  >\r\n                    <tbody>\r\n                      <tr>\r\n                        <td style="\r\n                          font-family: Arial, sans-serif; /* change font family */\r\n                          font-size: 16px; /* change font size */\r\n                          color: #ffffff; /* white text color */\r\n                          text-decoration: none; /* remove underline */\r\n                        ">\r\n                          Try it Now\r\n                        </td>\r\n                      </tr>\r\n                    </tbody>\r\n                  </table>\r\n                </a>\r\n                <div class="">\r\n                  <p\r\n                    class="text-gray-700"\r\n                    style="\r\n                      margin-top: 15px !important;\r\n                      line-height: 24px;\r\n                      font-size: 13px;\r\n                      color: #4a5568;\r\n                      width: 100%;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                  >\r\n                    Any questions, please reply to this email, our team is\r\n                    always happy to help.\r\n                  </p>\r\n                </div>\r\n              </td>\r\n            </tr>\r\n          </tbody>\r\n        </table>\r\n      </td>\r\n    </tr>\r\n  </tbody>\r\n</table>\r\n\r\n\r\n                            <table\r\n                              class="s-10 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 40px;\r\n                                      font-size: 40px;\r\n                                      width: 100%;\r\n                                      height: 40px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="40"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div>\r\n                              <table\r\n                                class="s-2 w-full"\r\n                                role="presentation"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="width: 100%"\r\n                                width="100%"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 8px;\r\n                                        font-size: 8px;\r\n                                        width: 100%;\r\n                                        height: 8px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                      width="100%"\r\n                                      height="8"\r\n                                    >\r\n                                      &#160;\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                              <table\r\n                                class="ax-center"\r\n                                role="presentation"\r\n                                align="center"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="margin: 0 auto"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 24px;\r\n                                        font-size: 16px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                    >\r\n                                      <img\r\n                                        class="w-48"\r\n                                        src="https://avian.io/static/images/avian_logo_full.png"\r\n                                        alt="Avian Logo"\r\n                                        style="\r\n                                          height: auto;\r\n                                          line-height: 100%;\r\n                                          outline: none;\r\n                                          text-decoration: none;\r\n                                          display: block;\r\n                                          width: 96px;\r\n                                          border-style: none;\r\n                                          border-width: 0;\r\n                                        "\r\n                                      />\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                              <table\r\n                                class="s-2 w-full"\r\n                                role="presentation"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="width: 100%"\r\n                                width="100%"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 8px;\r\n                                        font-size: 8px;\r\n                                        width: 100%;\r\n                                        height: 8px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                      width="100%"\r\n                                      height="8"\r\n                                    >\r\n                                      &#160;\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div\r\n                              class="text-muted text-center"\r\n                              style="color: #718096"\r\n                              align="center"\r\n                            >\r\n                              Sent from Avian.<br />\r\n                              315 W 36th St. 5th floor<br />\r\n                              New York, NY 10018, United States<br />\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                   &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div\r\n                            class="text-muted text-center"\r\n                            style="color: #718096"\r\n                            align="center"\r\n                          >\r\n                          </div>\r\n                          </td>\r\n                        </tr>\r\n                      </tbody>\r\n                    </table>\r\n                    <!--[if (gte mso 9)|(IE)]>\r\n                    </td>\r\n                  </tr>\r\n                </tbody>\r\n              </table>\r\n                    <![endif]-->\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n            <table\r\n              class="s-8 w-full"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n              width="100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    style="\r\n                      line-height: 32px;\r\n                      font-size: 32px;\r\n                      width: 100%;\r\n                      height: 32px;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                    width="100%"\r\n                    height="32"\r\n                  >\r\n                    &#160;\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n          </td>\r\n        </tr>\r\n      </tbody>\r\n    </table>\r\n  <p>If you&#39;d like to unsubscribe and stop receiving these emails <a href="http://url3045.avian.io/wf/unsubscribe?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIhPCFEthoZsRaCIVhjsUvwtABkdlPLsuPJkv82ZuH3p6bNwUpkB1XPMHoDEEdNcw63LEqUUlXh8a5DzB42TYmKZebnTD3YcsC-2BtaOaTgVnTyzpEQUO93kDk0RxpfZCpuiMaZw3S4csOdNVY1b-2BYs7pWL3TWtc-2BWqHgk71Ej4GZcj6DMaBWLx9KZri2OcNOdeR4-3D"> click here </a>.</p>\r\n<img src="http://url3045.avian.io/wf/open?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIhPCFEthoZsRaCIVhjsUvwtF26ehxTrxJCmFhvrhMQHZQFB74avW3RS4CHGQzA3PStcTbUrXO-2F2T31bdY-2BI4f8jLnFVr6OyyQyVJ3qjjKZ9RfPkNMebXeAhNDNRH2xOVxLyoGhZv89vAvjNUi7dULIgbOUXMaNyuIog6g6bMdHqQA-3D-3D" alt="" width="1" height="1" border="0" style="height:1px !important;width:1px !important;border-width:0 !important;margin-top:0 !important;margin-bottom:0 !important;margin-right:0 !important;margin-left:0 !important;padding-top:0 !important;padding-bottom:0 !important;padding-right:0 !important;padding-left:0 !important;"/></body>\r\n</html>\r\n	\N	t	f	f	0	\N	2024-12-10 17:06:30+00	2025-08-15 23:31:02.887131+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.887131+00	2025-08-15 23:31:23.913568+00
d88820df-4a5d-4751-ba30-b6b1c62357b9	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	b9780ba8-90c1-40d0-a459-b1d39b3682ea@featuredforgellc.com	<a2d747b3-bdb7-4813-bf3c-36b3e110d519@featuredforgellc.com>	\N	support@snorkell.ai	Junaid  Akram <jud@rgellc.com>	Junaid  Akram	{support@snorkell.ai}	\N	\N	\N	Re: RE: Let's collaborate!	Did you checked my previous email?\r\n\r\nIf not, you are missing out.\r\n\r\n\r\nDon't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3\r\n\r\nOn Wed, July 24, 2024 2:48 PM, Junaid  Akram <junaid@featuredforgellc.com>\r\n[junaid@featuredforgellc.com]> wrote:\r\n\r\n> One of our clients, James, saw a 93 percent boost in growth after getting featured on major news outlets like Yahoo, MSN, and Business Insider. Imagine what that could mean for your business.\r\n> \r\n> \r\n> Don't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3\r\n> On Tue, July 16, 2024 2:10 PM, Junaid  Akram <junaid@featuredforgellc.com>\r\n> [junaid@featuredforgellc.com]> wrote:\r\n> \r\n> > Hi,\r\n> > \r\n> > I came across your website and it's perfect.\r\n> > \r\n> > We can publish your story on MSN, Yahoo, Business Insider, Newsmax, Benzinga and 300+ other news sites, reaching 228m+ monthly visitors.\r\n> > \r\n> > Interested? I can share some recent case studies.\r\n> > \r\n> > Regards,\r\n> > Junaid\r\n> > \r\n> > \r\n> > Don't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3\r\n> > \r\n> >\r\n\r\npicture [https://inst.featuredforgellc.com/tmid_a/TuZUnXmMH2oOwRoVSyxNd] logo [https://inst.featuredforgellc.com/tmid_a/TuZUnXmMH2oOwRoVSyxNd]	<div>Did you checked my previous email?</div><div><br></div><div>If not, you are missing out.</div><div><br></div><div><br style="box-sizing: border-box; font-family: Averta, sans-serif;"><span style="font-size: 12px;">Don't want me to contact you again? <a href="https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3" target="_blank">Click here</a></span></div>\r\n<div class="gmail_quote">\r\n   On Wed, July 24, 2024 2:48 PM, Junaid  Akram <span dir="ltr">&lt;<a href="mailto:junaid@featuredforgellc.com" target="_blank">junaid@featuredforgellc.com</a>&gt;</span> wrote:<br>\r\n  <blockquote class="gmail_quote" style="margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex">\r\n    <div dir="ltr">\r\n    <div style="box-sizing: border-box;"><div>One of our clients, James, saw a 93 percent boost in growth after getting featured on major news outlets like Yahoo, MSN, and Business Insider. Imagine what that could mean for your business.</div><div><br></div><div><br></div><div><span style="font-size: 12px;">Don't want me to contact you again? <a href="https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3" target="_blank">Click here</a></span></div></div>\r\n      <div class="gmail_quote">\r\n         On Tue, July 16, 2024 2:10 PM, Junaid  Akram <span dir="ltr">&lt;<a href="mailto:junaid@featuredforgellc.com" target="_blank">junaid@featuredforgellc.com</a>&gt;</span> wrote:<br>\r\n        <blockquote class="gmail_quote" style="margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex">\r\n          <div dir="ltr">\r\n          <div style="box-sizing: border-box;"><div >Hi,</div><div><br></div><div>I came across your website and it's perfect.</div><div><br></div><div>We can publish your story on MSN, Yahoo, Business Insider, Newsmax, Benzinga and 300+ other news sites, reaching 228m+ monthly visitors.</div><div><br></div><div>Interested? I can share some recent case studies.</div><div><br></div><div>Regards,</div><div>Junaid</div><div><br></div><div><br><span style="font-size:12px;">Don't want me to contact you again?&nbsp;</span><a href="https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3" target="_blank"><span style="font-size:12px;">Click here</span></a><span style="font-size:12px;"><br></span></div></div>\r\n          </div>\r\n        </blockquote>\r\n      </div>\r\n    </div>\r\n  </blockquote>\r\n</div>\r\n<img alt="" src="https://inst.featuredforgellc.com/tmid_a/TuZUnXmMH2oOwRoVSyxNd">	Did you checked my previous email? If not, you are missing out. Don't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub...	t	t	f	0	\N	2024-07-29 14:37:19+00	2025-08-15 23:31:02.846206+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:02.846206+00	2025-08-15 23:31:29.460213+00
1608d6eb-bc1e-4322-8473-20406e49c4bc	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAOne=D56Z0R8-2HfxH-EznhO2yCqzMy2dOGR+hvaCCH1cdHNew@mail.gmail.com	\N	\N	support@snorkell.ai	Suman Saurabh <sumanrocs@gmail.com>	Suman Saurabh	{support@snorkell.ai}	\N	\N	\N	test	test\r\n	<div dir="ltr">test</div>\r\n	test	t	f	f	0	\N	2025-08-15 06:11:44+00	2025-08-15 23:31:01.927122+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.927122+00	2025-08-15 23:31:10.330897+00
d03bb3eb-830e-4da6-a1cf-f3908794cf4f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAAxuJtEH=eVSx_zT723vzAg8+k=cbs_z0WBxDaW98a8Gqwc8ZQ@mail.gmail.com	<CAAxuJtGhzHC4LBaW7AgEzFHKz8go9ZQoYk7Pd6TLWhzjBPgj0A@mail.gmail.com>	\N	support@snorkell.ai	Dee - Founder <dee@pearllemongroup.uk>	Dee - Founder	{support@snorkell.ai}	\N	\N	\N	Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell	 Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell\r\n\r\nHi team Snorkell,\r\n\r\nJust making one last attempt to reach out about offering a testimonial or\r\ncase study for Snorkell. Our team at Pearl Lemon has really enjoyed using\r\nit, and we’d love to help you highlight how awesome it is.\r\n\r\nIf that sounds good, let me know—we’re more than happy to support!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*Speak soon,DeeBalding 38-year-old British-Indian, living in the Italian\r\ncountryside. Oreo milkshake addict, ultra-marathon runner (but super slow),\r\nproud cat parent to Jenny, and the guy with more tattoos than push-ups\r\n(still working on getting past 10 a day)! :)Unsubscribe\r\n<http://w1.mssnrp.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/319ecb63-8afc-48fa-8877-330f698c0dc7>*\r\n\r\n\r\nOn Thu, Nov 28, 2024 at 9:51 AM "Dee - Founder" <dee@pearllemongroup.uk>\r\nwrote:\r\n\r\nHi team Snorkell,\r\n\r\nHope all’s going well! Just following up on my last email—some of our team\r\nat Pearl Lemon have been loving Snorkell in their daily workflow. It’s been\r\na game-changer!\r\n\r\nWe’d love to help by offering a testimonial or even putting together a case\r\nstudy if that would be useful.\r\n\r\nLet me know if you’re interested!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*Thanks so much,Dee38, balding, British-Indian, and living the dream in the\r\nItalian countryside. I’m hooked on Oreo milkshakes, run ultramarathons\r\n(though turtles could probably outpace me), and my cat Jenny is the real\r\nboss of the house. Oh, and I’ve got more tattoos than common sense, but I’m\r\nstill struggling to hit 10 push-ups a day :)Unsubscribe\r\n<http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441>*\r\n\r\n\r\nOn Tue, Nov 19, 2024 at 4:02 PM "Dee - Founder" <dee@pearllemongroup.uk>\r\nwrote:\r\n\r\nHey team Snorkell,\r\n\r\nDee here, founder of Pearl Lemon Group.\r\n\r\nA few of our team members have recently started using Snorkell as part of\r\ntheir workflow, and they’re loving it! :)\r\n\r\nJust wanted to say—awesome job! We know how important testimonials and case\r\nstudies can be, especially when you’re building something great.\r\n\r\nAs fellow founders, we’d love to support you. If there’s any way we can\r\nhelp out with that, just give us a shout!\r\n\r\nThanks a ton!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*DeeDodgy marathon runner, balding 38-year-old with far too many\r\nregrettable tattoos :)Unsubscribe\r\n<http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/3928a83c-91b6-48e4-906c-496c60707e61>*\r\n	\r\n      <html>\r\n      <head>\r\n        <title>Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell</title>\r\n        <meta content="text/html;" charset="utf-8" http-equiv="Content-Type">\r\n      </head>\r\n      <body><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hi team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Just making one last attempt to reach out about offering a testimonial or case study for Snorkell. Our team at Pearl Lemon has really enjoyed using it, and we’d love to help you highlight how awesome it is.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">If that sounds good, let me know—we’re more than happy to support!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-c7157281-7fff-0bda-c3a5-597b987dce95" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Speak soon,<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee</span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Balding 38-year-old British-Indian, living in the Italian countryside. Oreo milkshake addict, ultra-marathon runner (but super slow), proud cat parent to Jenny, and the guy with more tattoos than push-ups (still working on getting past 10 a day)! :)<br><br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssnrp.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/319ecb63-8afc-48fa-8877-330f698c0dc7">Unsubscribe</a><br></span></strong></p><img alt="" width="1" height="1" class="beacon-o" src="http://w1.mssnrp.com/prod/open/319ecb63-8afc-48fa-8877-330f698c0dc7" style="float:left;margin-left:-1px;position:absolute;"><div class="reply-chain"><br><br>On Thu, Nov 28, 2024 at 9:51 AM &quot;Dee - Founder&quot; &lt;<a href="mailto:dee@pearllemongroup.uk">dee@pearllemongroup.uk</a>&gt; wrote:<br><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hi team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hope all’s going well! Just following up on my last email—some of our team at Pearl Lemon have been loving Snorkell in their daily workflow. It’s been a game-changer!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">We’d love to help by offering a testimonial or even putting together a case study if that would be useful.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Let me know if you’re interested!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-79d5e964-7fff-446d-c9fd-c340abd7873c" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Thanks so much,<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">38, balding, British-Indian, and living the dream in the Italian countryside. I’m hooked on Oreo milkshakes, run ultramarathons (though turtles could probably outpace me), and my cat Jenny is the real boss of the house. Oh, and I’ve got more tattoos than common sense, but I’m still struggling to hit 10 push-ups a day :)<br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441">Unsubscribe</a><br></span></strong></p>\r\n<strong id="docs-internal-guid-ec92e9cf-7fff-2bd2-c924-ec227ae7e30c" style="font-weight:normal"></strong><br><br>On Tue, Nov 19, 2024 at 4:02 PM &quot;Dee - Founder&quot; &lt;<a href="mailto:dee@pearllemongroup.uk">dee@pearllemongroup.uk</a>&gt; wrote:<br><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hey team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee here, founder of Pearl Lemon Group.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">A few of our team members have recently started using Snorkell as part of their workflow, and they’re loving it! :)</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Just wanted to say—awesome job! We know how important testimonials and case studies can be, especially when you’re building something great.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">As fellow founders, we’d love to support you. If there’s any way we can help out with that, just give us a shout!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Thanks a ton!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-ed630042-7fff-5576-cc31-593961249aed" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dodgy marathon runner, balding 38-year-old with far too many regrettable tattoos :)<br><br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/3928a83c-91b6-48e4-906c-496c60707e61">Unsubscribe</a><br></span></strong></p>\r\n<strong id="docs-internal-guid-5c91a878-7fff-081f-784b-47622bbf035f" style="font-weight:normal"></strong></div></body>\r\n      </html>\r\n	Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell Hi team Snorkell, Just making one last attempt to reach out about offering a testimonial ...	t	t	f	0	\N	2024-12-03 10:13:58+00	2025-08-15 23:30:42.74376+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.74376+00	2025-08-15 23:32:21.114283+00
e9e61cd3-e9c6-44e7-81db-afcc116cb855	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	MEYPR01MB7118E81AB4552C7755BD9D8CD94F2@MEYPR01MB7118.ausprd01.prod.outlook.com	\N	\N	support@snorkell.ai	Alistair <Alistair.y@f.com.au>	Alistair Rigney	{"support@snorkell.ai <support@snorkell.ai>"}	\N	\N	\N	Could we use snorkell for an Azure Dev Ops repo	Dear Support,\r\n                           Guys love your work, but could we use it with Azure Dev Ops Repos?\r\nBest Regards,\r\n                            Alistair\r\n\r\nDisclaimer\r\n\r\nThe information contained in this communication from the sender is confidential. It is intended solely for use by the recipient and others authorized to receive it. If you are not the recipient, you are hereby notified that any disclosure, copying, distribution or taking action in relation of the contents of this information is strictly prohibited and may be unlawful.\r\n\r\nThis email has been scanned for viruses and malware, and may have been automatically archived by Mimecast Ltd, an innovator in Software as a Service (SaaS) for business. Providing a safer and more useful place for your human generated data. Specializing in; Security, archiving and compliance. To find out more visit the Mimecast website.\r\n	<html><head>\r\n<meta http-equiv="Content-Type" content="text/html; charset=us-ascii">\r\n<meta name="Generator" content="Microsoft Word 15 (filtered medium)">\r\n<style><!--\r\n/* Font Definitions */\r\n@font-face\r\n\t{font-family:"Cambria Math";\r\n\tpanose-1:2 4 5 3 5 4 6 3 2 4;}\r\n@font-face\r\n\t{font-family:Aptos;}\r\n/* Style Definitions */\r\np.MsoNormal, li.MsoNormal, div.MsoNormal\r\n\t{margin:0cm;\r\n\tfont-size:12.0pt;\r\n\tfont-family:"Aptos",sans-serif;\r\n\tmso-ligatures:standardcontextual;\r\n\tmso-fareast-language:EN-US;}\r\nspan.EmailStyle17\r\n\t{mso-style-type:personal-compose;\r\n\tfont-family:"Aptos",sans-serif;\r\n\tcolor:windowtext;}\r\n.MsoChpDefault\r\n\t{mso-style-type:export-only;\r\n\tmso-fareast-language:EN-US;}\r\n@page WordSection1\r\n\t{size:612.0pt 792.0pt;\r\n\tmargin:72.0pt 72.0pt 72.0pt 72.0pt;}\r\ndiv.WordSection1\r\n\t{page:WordSection1;}\r\n--></style><!--[if gte mso 9]><xml>\r\n<o:shapedefaults v:ext="edit" spidmax="1026" />\r\n</xml><![endif]--><!--[if gte mso 9]><xml>\r\n<o:shapelayout v:ext="edit">\r\n<o:idmap v:ext="edit" data="1" />\r\n</o:shapelayout></xml><![endif]-->\r\n<style type="text/css">.style1 {font-family: "Times New Roman";}</style></head><body lang="EN-AU" link="#467886" vlink="#96607D" style="word-wrap:break-word">\r\n<div class="WordSection1">\r\n<p class="MsoNormal">Dear Support,<o:p></o:p></p>\r\n<p class="MsoNormal">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Guys love your work, but could we use it with Azure Dev Ops Repos?<o:p></o:p></p>\r\n<p class="MsoNormal">Best Regards,<o:p></o:p></p>\r\n<p class="MsoNormal">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp; Alistair<o:p></o:p></p>\r\n</div>\r\n\r\n\r\n<br><br><p style="font-family: Verdana; font-size:10pt; color:#666666;"><b>Disclaimer</b></p><p style="font-family: Verdana; font-size:8pt; color:#666666;">The information contained in this communication from the sender is confidential. It is intended solely for use by the recipient and others authorized to receive it. If you are not the recipient, you are hereby notified that any disclosure, copying, distribution or taking action in relation of the contents of this information is strictly prohibited and may be unlawful.<br><br>This email has been scanned for viruses and malware, and may have been automatically archived by <b>Mimecast Ltd</b>, an innovator in Software as a Service (SaaS) for business.  Providing a <b>safer</b> and <b>more useful</b> place for your human generated data.  Specializing in; Security, archiving and compliance. To find out more <a href="http://www.mimecast.com/products/" target="_blank">Click Here</a>.</p></body></html>\r\n	Dear Support, Guys love your work, but could we use it with Azure Dev Ops Repos? Best Regards, Alistair Disclaimer The information contained in this c...	t	f	f	0	\N	2024-02-13 04:59:56+00	2025-08-15 23:31:00.584681+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.584681+00	2025-08-16 00:17:31.992102+00
9f29fc69-3fe2-45ef-a835-7e4079d969a6	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	adc76df8-ea6e-49ab-b195-0b13c67bece0@saasipedia.com	\N	\N	support@snorkell.ai	Elij <elijah@saedia.com>	Elijah Parker	{support@snorkell.ai}	\N	\N	\N	RE: Partnership Opportunity!	Hi, \r\n\r\nI recently went through your Product Hunt launch and wanted to congratulate you.\r\n\r\nBut you can Submit your Saas on 200+ directories other than Product Hunt, IndieHackers, etc. and generate upto 50,000 organic visitors to your website.\r\n\r\nIt has helped a lot of SaaS startups like yours to get first 1,000 paying customers organically.\r\n\r\nInterested? I can share real-life case studies.\r\n\r\nRegards,\r\nElijah\r\n\r\n\r\nDon't want me to contact you again? Click here https://inst.saasipedia.com/unsub/1/47346fd9-605e-431f-bbd7-50c52f9594e8\r\n\r\n\r\npicture [https://inst.saasipedia.com/tmid_a/VnFjZ4trbEy41lfYvgNUy] logo [https://inst.saasipedia.com/tmid_a/VnFjZ4trbEy41lfYvgNUy]	<div>Hi,&nbsp;</div><div><br>I recently went through your Product Hunt launch and wanted to congratulate you.</div><div><br></div><div>But you can Submit your Saas on 200+ directories other than Product Hunt, IndieHackers, etc. and generate upto 50,000 organic visitors to your website.</div><div><br>It has helped a lot of SaaS startups like yours to get first 1,000 paying customers organically.</div><div><br>Interested? I can share real-life case studies.</div><div><br>Regards,</div><div>Elijah</div><div><br></div><div><br><span style="font-size: 12px;">Don't want me to contact you again?&nbsp;</span><a href="https://inst.saasipedia.com/unsub/1/47346fd9-605e-431f-bbd7-50c52f9594e8" target="_blank"><span style="font-size: 12px;">Click here</span></a><span style="font-size: 12px;"><br></span></div>\r\n<img alt="" src="https://inst.saasipedia.com/tmid_a/VnFjZ4trbEy41lfYvgNUy">	Hi,  I recently went through your Product Hunt launch and wanted to congratulate you. But you can Submit your Saas on 200+ directories other than Pro...	f	t	f	0	\N	2024-04-15 16:37:04+00	2025-08-15 23:30:33.842947+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.842947+00	2025-08-15 23:30:33.842947+00
87c6bea6-2088-4664-a66a-6f84a3da9da1	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	a2d747b3-bdb7-4813-bf3c-36b3e110d519@featuredforgellc.com	<b300f30c-f336-4731-b59c-6f6790aca543@featuredforgellc.com>	\N	support@snorkell.ai	Junaid <jaid@featorgellc.com>	Junaid  Akram	{support@snorkell.ai}	\N	\N	\N	Re: RE: Let's collaborate!	One of our clients, James, saw a 93 percent boost in growth after getting featured on major news outlets like Yahoo, MSN, and Business Insider. Imagine what that could mean for your business.\r\n\r\n\r\nDon't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3\r\n\r\nOn Tue, July 16, 2024 2:10 PM, Junaid  Akram <junaid@featuredforgellc.com>\r\n[junaid@featuredforgellc.com]> wrote:\r\n\r\n> Hi,\r\n> \r\n> I came across your website and it's perfect.\r\n> \r\n> We can publish your story on MSN, Yahoo, Business Insider, Newsmax, Benzinga and 300+ other news sites, reaching 228m+ monthly visitors.\r\n> \r\n> Interested? I can share some recent case studies.\r\n> \r\n> Regards,\r\n> Junaid\r\n> \r\n> \r\n> Don't want me to contact you again? Click here https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3\r\n> \r\n>\r\n\r\npicture [https://inst.featuredforgellc.com/tmid_a/9OZyjI-aUBoJ3XaDg-jVb] logo [https://inst.featuredforgellc.com/tmid_a/9OZyjI-aUBoJ3XaDg-jVb]	<div>One of our clients, James, saw a 93 percent boost in growth after getting featured on major news outlets like Yahoo, MSN, and Business Insider. Imagine what that could mean for your business.</div><div><br></div><div><br></div><div><span style="font-size: 12px;">Don't want me to contact you again? <a href="https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3" target="_blank">Click here</a></span></div>\r\n<div class="gmail_quote">\r\n   On Tue, July 16, 2024 2:10 PM, Junaid  Akram <span dir="ltr">&lt;<a href="mailto:junaid@featuredforgellc.com" target="_blank">junaid@featuredforgellc.com</a>&gt;</span> wrote:<br>\r\n  <blockquote class="gmail_quote" style="margin:0 0 0 .8ex;border-left:1px #ccc solid;padding-left:1ex">\r\n    <div dir="ltr">\r\n    <div style="box-sizing: border-box;"><div >Hi,</div><div><br></div><div>I came across your website and it's perfect.</div><div><br></div><div>We can publish your story on MSN, Yahoo, Business Insider, Newsmax, Benzinga and 300+ other news sites, reaching 228m+ monthly visitors.</div><div><br></div><div>Interested? I can share some recent case studies.</div><div><br></div><div>Regards,</div><div>Junaid</div><div><br></div><div><br><span style="font-size:12px;">Don't want me to contact you again?&nbsp;</span><a href="https://inst.featuredforgellc.com/unsub/1/8f6b460f-d1a0-45ac-810d-f94efd4238a3" target="_blank"><span style="font-size:12px;">Click here</span></a><span style="font-size:12px;"><br></span></div></div>\r\n    </div>\r\n  </blockquote>\r\n</div>\r\n<img alt="" src="https://inst.featuredforgellc.com/tmid_a/9OZyjI-aUBoJ3XaDg-jVb">	One of our clients, James, saw a 93 percent boost in growth after getting featured on major news outlets like Yahoo, MSN, and Business Insider. Imagin...	t	t	f	0	\N	2024-07-24 14:48:46+00	2025-08-15 23:30:42.737259+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.737259+00	2025-08-18 07:56:50.028857+00
0f1fb790-306a-4019-80dc-23c6d27bb85b	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	PkcYj_zHSQGytIZ5UDbzjA@geopod-ismtpd-1	\N	\N	support@snorkell.ai	Ana from Avian.io <info@avian.io>	Ana from Avian.io	{support@snorkell.ai}	\N	\N	\N	World's Fastest: Llama 405B up to 142 tok/s on Nvidia H200 with Avian.io API		<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.0 Transitional//EN" "http://www.w3.org/TR/REC-html40/loose.dtd">\r\n<html>\r\n  <head>\r\n    <meta http-equiv="x-ua-compatible" content="ie=edge" />\r\n    <meta name="x-apple-disable-message-reformatting" />\r\n    <meta name="viewport" content="width=device-width, initial-scale=1" />\r\n    <meta\r\n      name="format-detection"\r\n      content="telephone=no, date=no, address=no, email=no"\r\n    />\r\n    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />\r\n    <style type="text/css">\r\n      body,\r\n      table,\r\n      td {\r\n        font-family: Helvetica, Arial, sans-serif !important;\r\n      }\r\n      .ExternalClass {\r\n        width: 100%;\r\n      }\r\n      .ExternalClass,\r\n      .ExternalClass p,\r\n      .ExternalClass span,\r\n      .ExternalClass font,\r\n      .ExternalClass td,\r\n      .ExternalClass div {\r\n        line-height: 150%;\r\n      }\r\n      a {\r\n        text-decoration: none;\r\n      }\r\n      * {\r\n        color: inherit;\r\n      }\r\n      a[x-apple-data-detectors],\r\n      u + #body a,\r\n      #MessageViewBody a {\r\n        color: inherit;\r\n        text-decoration: none;\r\n        font-size: inherit;\r\n        font-family: inherit;\r\n        font-weight: inherit;\r\n        line-height: inherit;\r\n      }\r\n      img {\r\n        -ms-interpolation-mode: bicubic;\r\n      }\r\n      table:not([class^="s-"]) {\r\n        font-family: Helvetica, Arial, sans-serif;\r\n        mso-table-lspace: 0pt;\r\n        mso-table-rspace: 0pt;\r\n        border-spacing: 0px;\r\n        border-collapse: collapse;\r\n      }\r\n      table:not([class^="s-"]) td {\r\n        border-spacing: 0px;\r\n        border-collapse: collapse;\r\n      }\r\n      @media screen and (max-width: 600px) {\r\n        .w-full,\r\n        .w-full > tbody > tr > td {\r\n          width: 100% !important;\r\n        }\r\n        .w-12,\r\n        .w-12 > tbody > tr > td {\r\n          width: 48px !important;\r\n        }\r\n        .w-48,\r\n        .w-48 > tbody > tr > td {\r\n          width: 192px !important;\r\n        }\r\n        .pt-5:not(table),\r\n        .pt-5:not(.btn) > tbody > tr > td,\r\n        .pt-5.btn td a,\r\n        .py-5:not(table),\r\n        .py-5:not(.btn) > tbody > tr > td,\r\n        .py-5.btn td a {\r\n          padding-top: 20px !important;\r\n        }\r\n        .pb-5:not(table),\r\n        .pb-5:not(.btn) > tbody > tr > td,\r\n        .pb-5.btn td a,\r\n        .py-5:not(table),\r\n        .py-5:not(.btn) > tbody > tr > td,\r\n        .py-5.btn td a {\r\n          padding-bottom: 20px !important;\r\n        }\r\n        *[class*="s-lg-"] > tbody > tr > td {\r\n          font-size: 0 !important;\r\n          line-height: 0 !important;\r\n          height: 0 !important;\r\n        }\r\n        .s-2 > tbody > tr > td {\r\n          font-size: 8px !important;\r\n          line-height: 8px !important;\r\n          height: 8px !important;\r\n        }\r\n        .s-6 > tbody > tr > td {\r\n          font-size: 24px !important;\r\n          line-height: 24px !important;\r\n          height: 24px !important;\r\n        }\r\n        .s-8 > tbody > tr > td {\r\n          font-size: 32px !important;\r\n          line-height: 32px !important;\r\n          height: 32px !important;\r\n        }\r\n        .s-10 > tbody > tr > td {\r\n          font-size: 40px !important;\r\n          line-height: 40px !important;\r\n          height: 40px !important;\r\n        }\r\n      }\r\n    </style>\r\n  </head>\r\n  <body\r\n    class="bg-light"\r\n    style="\r\n      outline: 0;\r\n      width: 100%;\r\n      min-width: 100%;\r\n      height: 100%;\r\n      -webkit-text-size-adjust: 100%;\r\n      -ms-text-size-adjust: 100%;\r\n      font-family: Helvetica, Arial, sans-serif;\r\n      line-height: 24px;\r\n      font-weight: normal;\r\n      font-size: 16px;\r\n      -moz-box-sizing: border-box;\r\n      -webkit-box-sizing: border-box;\r\n      box-sizing: border-box;\r\n      color: #000000;\r\n      margin: 0;\r\n      padding: 0;\r\n      border-width: 0;\r\n    "\r\n    bgcolor="#f7fafc"\r\n  >\r\n    <table\r\n      class="bg-light body"\r\n      valign="top"\r\n      role="presentation"\r\n      border="0"\r\n      cellpadding="0"\r\n      cellspacing="0"\r\n      style="\r\n        outline: 0;\r\n        width: 100%;\r\n        min-width: 100%;\r\n        height: 100%;\r\n        -webkit-text-size-adjust: 100%;\r\n        -ms-text-size-adjust: 100%;\r\n        font-family: Helvetica, Arial, sans-serif;\r\n        line-height: 24px;\r\n        font-weight: normal;\r\n        font-size: 16px;\r\n        -moz-box-sizing: border-box;\r\n        -webkit-box-sizing: border-box;\r\n        box-sizing: border-box;\r\n        color: #000000;\r\n        margin: 0;\r\n        padding: 0;\r\n        border-width: 0;\r\n      "\r\n      bgcolor="#f7fafc"\r\n    >\r\n      <tbody>\r\n        <tr>\r\n          <td\r\n            valign="top"\r\n            style="line-height: 24px; font-size: 16px; margin: 0"\r\n            align="left"\r\n            bgcolor="#f7fafc"\r\n          >\r\n            <table\r\n              class="s-8 w-full"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n              width="100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    style="\r\n                      line-height: 32px;\r\n                      font-size: 32px;\r\n                      width: 100%;\r\n                      height: 32px;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                    width="100%"\r\n                    height="32"\r\n                  >\r\n                    &#160;\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n            <table\r\n              class="container"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    align="center"\r\n                    style="\r\n                      line-height: 24px;\r\n                      font-size: 16px;\r\n                      margin: 0;\r\n                      padding: 0 16px;\r\n                    "\r\n                  >\r\n                    <!--[if (gte mso 9)|(IE)]>\r\n                      <table align="center" role="presentation">\r\n                        <tbody>\r\n                          <tr>\r\n                            <td width="600">\r\n                    <![endif]-->\r\n                    <table\r\n                      align="center"\r\n                      role="presentation"\r\n                      border="0"\r\n                      cellpadding="0"\r\n                      cellspacing="0"\r\n                      style="width: 100%; max-width: 600px; margin: 0 auto"\r\n                    >\r\n                      <tbody>\r\n                        <tr>\r\n                          <td\r\n                            style="\r\n                              line-height: 24px;\r\n                              font-size: 16px;\r\n                              margin: 0;\r\n                            "\r\n                            align="left"\r\n                          >\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div class="">\r\n                              <table\r\n                                class="ax-center"\r\n                                role="presentation"\r\n                                align="center"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="margin: 0 auto"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 24px;\r\n                                        font-size: 16px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                    >\r\n                                      <img\r\n                                        class="w-24"\r\n                                        src="https://avian.io/static/images/logomark.png"\r\n                                        alt="Avian Logo"\r\n                                        style="\r\n                                          height: auto;\r\n                                          line-height: 100%;\r\n                                          outline: none;\r\n                                          text-decoration: none;\r\n                                          display: block;\r\n                                          width: 48px;\r\n                                          border-style: none;\r\n                                          border-width: 0;\r\n                                        "\r\n                                        width="48"\r\n                                      />\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <table\r\n                              class="s-10 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 40px;\r\n                                      font-size: 40px;\r\n                                      width: 100%;\r\n                                      height: 40px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="40"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n\r\n                            \r\n\r\n<table\r\n  class="card"\r\n  role="presentation"\r\n  border="0"\r\n  cellpadding="0"\r\n  cellspacing="0"\r\n  style="\r\n    border-radius: 6px;\r\n    border-collapse: separate !important;\r\n    width: 100%;\r\n    overflow: hidden;\r\n    border: 1px solid #e2e8f0;\r\n  "\r\n  bgcolor="#ffffff"\r\n>\r\n  <tbody>\r\n    <tr>\r\n      <td\r\n        style="line-height: 24px; font-size: 16px; width: 100%; margin: 0"\r\n        align="left"\r\n        bgcolor="#ffffff"\r\n      >\r\n        <table\r\n          class="card-body"\r\n          role="presentation"\r\n          border="0"\r\n          cellpadding="0"\r\n          cellspacing="0"\r\n          style="width: 100%"\r\n        >\r\n          <tbody>\r\n            <tr>\r\n              <td\r\n                style="\r\n                  line-height: 24px;\r\n                  font-size: 16px;\r\n                  width: 100%;\r\n                  margin: 0;\r\n                  padding: 40px;\r\n                "\r\n                align="left"\r\n              >\r\n                <h1\r\n                  class="h3"\r\n                  style="\r\n                    padding-top: 0;\r\n                    padding-bottom: 0;\r\n                    font-weight: 500;\r\n                    vertical-align: baseline;\r\n                    font-size: 28px;\r\n                    line-height: 33.6px;\r\n                    margin: 0;\r\n                  "\r\n                  align="left"\r\n                >\r\n                World's Fastest: Llama 405B up to 142 tok/s on Nvidia H200 with Avian.io API</h1>\r\n                <table\r\n                  class="s-2 w-full"\r\n                  role="presentation"\r\n                  border="0"\r\n                  cellpadding="0"\r\n                  cellspacing="0"\r\n                  style="width: 100%"\r\n                  width="100%"\r\n                >\r\n                  <tbody>\r\n                    <tr>\r\n                      <td\r\n                        style="\r\n                          line-height: 8px;\r\n                          font-size: 8px;\r\n                          width: 100%;\r\n                          height: 8px;\r\n                          margin: 0;\r\n                        "\r\n                        align="left"\r\n                        width="100%"\r\n                        height="8"\r\n                      >\r\n                        &#160;\r\n                      </td>\r\n                    </tr>\r\n                  </tbody>\r\n                </table>\r\n                <table\r\n                  class="py-2"\r\n                  role="presentation"\r\n                  border="0"\r\n                  cellpadding="0"\r\n                  cellspacing="0"\r\n                >\r\n                  <tbody>\r\n                    <tr>\r\n                      <td\r\n                        style="\r\n                          line-height: 24px;\r\n                          font-size: 16px;\r\n                          padding-bottom: 20px;\r\n                          margin: 0;\r\n                        "\r\n                        align="left"\r\n                      >\r\n                        <div class="">\r\n                          <p\r\n                            class="text-gray-700"\r\n                            style="\r\n                              line-height: 24px;\r\n                              font-size: 20px;\r\n                              color: #4a5568;\r\n                              width: 100%;\r\n                              margin: 0;\r\n                            "\r\n                            align="left"\r\n                          ></p>\r\n                          <div><a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3DKlw0_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZgkD6y0FVhHcyH64Avto45FZdWtlTBTJpiKNhxz9AQEFztha4m6ZonaEO6FpYOZS0IUFvRHvpZSlQds6jzMQ8spZT2HA6rYt2oD-2B4Reui76P-2FI-2FqdS2aBUdPCMebiE97tQSSQBYcQC36rT3s2Q3ZCQOtDRYfL4d082aociTKBkXg-3D-3D"><img src="https://i.imgur.com/tttuyis.png" alt="Avian Llama 405B Performance Benchmark"></a> <div> <p><strong>Exciting news! Avian.io is thrilled to announce the official launch of our World Record Llama 405B, powered by Nvidia H200.</strong></p> <p> The Llama 405B model achieves up to 142 tok/s and is only the first of two World Record Tier 405B models we will be launching this year.</p> <p> With this, we have effectively launched a model that is twice the speed and half the price of OpenAI's flagship 4o model.</p><p> We've decided to launch at a competitive price of $3 per million tokens, and the model is integrated into all of Avian's system as of now.</p>\r\n\r\n\r\n                          <p>Please get in touch with any questions, or queries about our bespoke finetuning and inference stack than can speed up your LLMs by 5-20x: <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DMP0utklEKqkkbx1xPVUfEPeeFaNeotwZeJr2pawexzOIW7l_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZgkD6y0FVhHcyH64Avto45x0bwQymg-2FugupNdKAm7TshBHnaajtOnpWsQWxJf-2FO5Bw1YOfemkDm5j5JkGVHphhZtRW8bHxv4tK90qjKchvHCkZzGBLGwwZ1j9P4EKhpZ7tiVpud3XGfGrfHVBSuLanX8qYBIUIhNTmNLErYHyTPA-3D-3D" style="color: blue;"> Book a Chat With Us</a> </p>\r\n                            \r\n                            The Avian Team</p> </div>\r\n                      </td>\r\n                    </tr>\r\n                  </tbody>\r\n                </table>\r\n                <a href="http://url3045.avian.io/ls/click?upn=u001.SR9b1GsxrQzVySydzI87DAJk9QoW2My-2BTEnHbC0SaGU-3DiISs_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZgkD6y0FVhHcyH64Avto4589ppVfR6u8XUxNqnQwSs3GIpErR0X4inMJ65pig-2FeNzqRRjgExXkupmiK5-2BOPHJHUXm6W3wuvjsOCL9J9jy3Y4M-2B5YKYd9hohSVo7tH64kiZo2zxF-2FdVhhOqsH3wp7NydgTvlkqgdvJ5U9y3X7EUvw-3D-3D">\r\n                  <table\r\n                    class="btn btn-primary"\r\n                    role="presentation"\r\n                    border="0"\r\n                    cellpadding="0"\r\n                    cellspacing="0"\r\n                    style="\r\n                      border-radius: 6px;\r\n                      border-collapse: separate!important;\r\n                      background-color: #007bff; /* blue background color */\r\n                      padding: 10px 20px; /* add some padding */\r\n                      text-align: center; /* center the text */\r\n                      cursor: pointer; /* change cursor to pointer on hover */\r\n                    "\r\n                  >\r\n                    <tbody>\r\n                      <tr>\r\n                        <td style="\r\n                          font-family: Arial, sans-serif; /* change font family */\r\n                          font-size: 16px; /* change font size */\r\n                          color: #ffffff; /* white text color */\r\n                          text-decoration: none; /* remove underline */\r\n                        ">\r\n                          Try it Now\r\n                        </td>\r\n                      </tr>\r\n                    </tbody>\r\n                  </table>\r\n                </a>\r\n                <div class="">\r\n                  <p\r\n                    class="text-gray-700"\r\n                    style="\r\n                      margin-top: 15px !important;\r\n                      line-height: 24px;\r\n                      font-size: 13px;\r\n                      color: #4a5568;\r\n                      width: 100%;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                  >\r\n                    Any questions, please reply to this email, our team is\r\n                    always happy to help.\r\n                  </p>\r\n                </div>\r\n              </td>\r\n            </tr>\r\n          </tbody>\r\n        </table>\r\n      </td>\r\n    </tr>\r\n  </tbody>\r\n</table>\r\n\r\n\r\n                            <table\r\n                              class="s-10 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 40px;\r\n                                      font-size: 40px;\r\n                                      width: 100%;\r\n                                      height: 40px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="40"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div>\r\n                              <table\r\n                                class="s-2 w-full"\r\n                                role="presentation"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="width: 100%"\r\n                                width="100%"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 8px;\r\n                                        font-size: 8px;\r\n                                        width: 100%;\r\n                                        height: 8px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                      width="100%"\r\n                                      height="8"\r\n                                    >\r\n                                      &#160;\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                              <table\r\n                                class="ax-center"\r\n                                role="presentation"\r\n                                align="center"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="margin: 0 auto"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 24px;\r\n                                        font-size: 16px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                    >\r\n                                      <img\r\n                                        class="w-48"\r\n                                        src="https://avian.io/static/images/avian_logo_full.png"\r\n                                        alt="Avian Logo"\r\n                                        style="\r\n                                          height: auto;\r\n                                          line-height: 100%;\r\n                                          outline: none;\r\n                                          text-decoration: none;\r\n                                          display: block;\r\n                                          width: 96px;\r\n                                          border-style: none;\r\n                                          border-width: 0;\r\n                                        "\r\n                                      />\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                              <table\r\n                                class="s-2 w-full"\r\n                                role="presentation"\r\n                                border="0"\r\n                                cellpadding="0"\r\n                                cellspacing="0"\r\n                                style="width: 100%"\r\n                                width="100%"\r\n                              >\r\n                                <tbody>\r\n                                  <tr>\r\n                                    <td\r\n                                      style="\r\n                                        line-height: 8px;\r\n                                        font-size: 8px;\r\n                                        width: 100%;\r\n                                        height: 8px;\r\n                                        margin: 0;\r\n                                      "\r\n                                      align="left"\r\n                                      width="100%"\r\n                                      height="8"\r\n                                    >\r\n                                      &#160;\r\n                                    </td>\r\n                                  </tr>\r\n                                </tbody>\r\n                              </table>\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                    &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div\r\n                              class="text-muted text-center"\r\n                              style="color: #718096"\r\n                              align="center"\r\n                            >\r\n                              Sent from Avian.<br />\r\n                              315 W 36th St. 5th floor<br />\r\n                              New York, NY 10018, United States<br />\r\n                            </div>\r\n                            <table\r\n                              class="s-6 w-full"\r\n                              role="presentation"\r\n                              border="0"\r\n                              cellpadding="0"\r\n                              cellspacing="0"\r\n                              style="width: 100%"\r\n                              width="100%"\r\n                            >\r\n                              <tbody>\r\n                                <tr>\r\n                                  <td\r\n                                    style="\r\n                                      line-height: 24px;\r\n                                      font-size: 24px;\r\n                                      width: 100%;\r\n                                      height: 24px;\r\n                                      margin: 0;\r\n                                    "\r\n                                    align="left"\r\n                                    width="100%"\r\n                                    height="24"\r\n                                  >\r\n                                   &#160;\r\n                                  </td>\r\n                                </tr>\r\n                              </tbody>\r\n                            </table>\r\n                            <div\r\n                            class="text-muted text-center"\r\n                            style="color: #718096"\r\n                            align="center"\r\n                          >\r\n                          </div>\r\n                          </td>\r\n                        </tr>\r\n                      </tbody>\r\n                    </table>\r\n                    <!--[if (gte mso 9)|(IE)]>\r\n                    </td>\r\n                  </tr>\r\n                </tbody>\r\n              </table>\r\n                    <![endif]-->\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n            <table\r\n              class="s-8 w-full"\r\n              role="presentation"\r\n              border="0"\r\n              cellpadding="0"\r\n              cellspacing="0"\r\n              style="width: 100%"\r\n              width="100%"\r\n            >\r\n              <tbody>\r\n                <tr>\r\n                  <td\r\n                    style="\r\n                      line-height: 32px;\r\n                      font-size: 32px;\r\n                      width: 100%;\r\n                      height: 32px;\r\n                      margin: 0;\r\n                    "\r\n                    align="left"\r\n                    width="100%"\r\n                    height="32"\r\n                  >\r\n                    &#160;\r\n                  </td>\r\n                </tr>\r\n              </tbody>\r\n            </table>\r\n          </td>\r\n        </tr>\r\n      </tbody>\r\n    </table>\r\n  <p>If you&#39;d like to unsubscribe and stop receiving these emails <a href="http://url3045.avian.io/wf/unsubscribe?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIhByuc3a46tckQNdBoqaGs247CZ4vHb7oUCuBhatPono-2FHpt4uZDPJJs7aIWGjsHQasVkdAi397drauDuRTVRZoleMLx-2BUtQOnrPEJqD86Cy2Soz4NcZbxAqbJ5dFTssA1KBQ2bgxSlGbLF8Ae59KdBJ-2BVwqx1QcEYZVtG3JYZ6jtpMH85Wuc26BwHus8Cp7Yc-3D"> click here </a>.</p>\r\n<img src="http://url3045.avian.io/wf/open?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIhByuc3a46tckQNdBoqaGs2hUmSlnnLUM9bNZwsD6w88P1esacejk-2BSygXqySyarzTtd7Z4ksiEzERJ-2FT7-2BPlBXEl-2Ft6Z7Vmxuay2suEy8YNGwqqLD5jOjDT2m8wyGsIPKr5rBYkxqEhmnA5sUHzqJ2FQRB-2BziERyO8M-2FRj1Dp0Aw-3D-3D" alt="" width="1" height="1" border="0" style="height:1px !important;width:1px !important;border-width:0 !important;margin-top:0 !important;margin-bottom:0 !important;margin-right:0 !important;margin-left:0 !important;padding-top:0 !important;padding-bottom:0 !important;padding-right:0 !important;padding-left:0 !important;"/></body>\r\n</html>\r\n	\N	t	f	f	0	\N	2024-10-31 17:46:56+00	2025-08-15 23:30:33.801549+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.801549+00	2025-08-15 23:32:02.746393+00
832cbd16-89e0-45e1-8375-d0d8cba93190	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAAxuJtGhzHC4LBaW7AgEzFHKz8go9ZQoYk7Pd6TLWhzjBPgj0A@mail.gmail.com	<CAAxuJtEw5bcbku=-Nd-mnN5Qe_zomYgM1mSJt79jHOPCSwVJwQ@mail.gmail.com>	\N	support@snorkell.ai	Dee - Founder <dee@pearllemongroup.uk>	Dee - Founder	{support@snorkell.ai}	\N	\N	\N	Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell	 Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell\r\n\r\nHi team Snorkell,\r\n\r\nHope all’s going well! Just following up on my last email—some of our team\r\nat Pearl Lemon have been loving Snorkell in their daily workflow. It’s been\r\na game-changer!\r\n\r\nWe’d love to help by offering a testimonial or even putting together a case\r\nstudy if that would be useful.\r\n\r\nLet me know if you’re interested!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*Thanks so much,Dee38, balding, British-Indian, and living the dream in the\r\nItalian countryside. I’m hooked on Oreo milkshakes, run ultramarathons\r\n(though turtles could probably outpace me), and my cat Jenny is the real\r\nboss of the house. Oh, and I’ve got more tattoos than common sense, but I’m\r\nstill struggling to hit 10 push-ups a day :)Unsubscribe\r\n<http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441>*\r\n\r\n\r\nOn Tue, Nov 19, 2024 at 4:02 PM "Dee - Founder" <dee@pearllemongroup.uk>\r\nwrote:\r\n\r\nHey team Snorkell,\r\n\r\nDee here, founder of Pearl Lemon Group.\r\n\r\nA few of our team members have recently started using Snorkell as part of\r\ntheir workflow, and they’re loving it! :)\r\n\r\nJust wanted to say—awesome job! We know how important testimonials and case\r\nstudies can be, especially when you’re building something great.\r\n\r\nAs fellow founders, we’d love to support you. If there’s any way we can\r\nhelp out with that, just give us a shout!\r\n\r\nThanks a ton!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*DeeDodgy marathon runner, balding 38-year-old with far too many\r\nregrettable tattoos :)Unsubscribe\r\n<http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441/3928a83c-91b6-48e4-906c-496c60707e61>*\r\n	\r\n      <html>\r\n      <head>\r\n        <title>Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell</title>\r\n        <meta content="text/html;" charset="utf-8" http-equiv="Content-Type">\r\n      </head>\r\n      <body><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hi team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hope all’s going well! Just following up on my last email—some of our team at Pearl Lemon have been loving Snorkell in their daily workflow. It’s been a game-changer!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">We’d love to help by offering a testimonial or even putting together a case study if that would be useful.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Let me know if you’re interested!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-79d5e964-7fff-446d-c9fd-c340abd7873c" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Thanks so much,<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">38, balding, British-Indian, and living the dream in the Italian countryside. I’m hooked on Oreo milkshakes, run ultramarathons (though turtles could probably outpace me), and my cat Jenny is the real boss of the house. Oh, and I’ve got more tattoos than common sense, but I’m still struggling to hit 10 push-ups a day :)<br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441">Unsubscribe</a><br></span></strong></p>\r\n<strong id="docs-internal-guid-ec92e9cf-7fff-2bd2-c924-ec227ae7e30c" style="font-weight:normal"></strong><img alt="" width="1" height="1" class="beacon-o" src="http://w1.mssusw.com/prod/open/34c9b35f-666c-44c9-9c6c-415ce94c5441" style="float:left;margin-left:-1px;position:absolute;"><div class="reply-chain"><br><br>On Tue, Nov 19, 2024 at 4:02 PM &quot;Dee - Founder&quot; &lt;<a href="mailto:dee@pearllemongroup.uk">dee@pearllemongroup.uk</a>&gt; wrote:<br><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hey team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee here, founder of Pearl Lemon Group.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">A few of our team members have recently started using Snorkell as part of their workflow, and they’re loving it! :)</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Just wanted to say—awesome job! We know how important testimonials and case studies can be, especially when you’re building something great.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">As fellow founders, we’d love to support you. If there’s any way we can help out with that, just give us a shout!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Thanks a ton!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-ed630042-7fff-5576-cc31-593961249aed" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dodgy marathon runner, balding 38-year-old with far too many regrettable tattoos :)<br><br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/34c9b35f-666c-44c9-9c6c-415ce94c5441/3928a83c-91b6-48e4-906c-496c60707e61">Unsubscribe</a><br></span></strong></p>\r\n<strong id="docs-internal-guid-5c91a878-7fff-081f-784b-47622bbf035f" style="font-weight:normal"></strong></div></body>\r\n      </html>\r\n	Re: Exciting Testimonial Offer from Pearl Lemon for Snorkell Hi team Snorkell, Hope all’s going well! Just following up on my last email—some of o...	t	t	f	0	\N	2024-11-28 09:51:58+00	2025-08-15 23:30:33.85694+00	synced	\N	\N	t	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.85694+00	2025-08-15 23:56:13.520571+00
1af8e96d-f010-4c82-8ba2-05aed4953278	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	ad4b69e9-70ef-4ab3-9eae-39756e32abff@us-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@us-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: 6180ca4f1e8d4c84c1e2eb81a49fbb10924319044c9e035e1768de605234bea7			\N	f	f	f	0	\N	2024-02-21 00:01:24+00	2025-08-15 23:30:33.817804+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.817804+00	2025-08-15 23:30:33.817804+00
ea3b277e-019a-4f00-b1f2-63faaf53e0ab	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	11192426104051859693@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 11192426104051859693			\N	f	f	f	0	\N	2024-03-01 23:59:59+00	2025-08-15 23:30:33.826627+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.826627+00	2025-08-15 23:30:33.826627+00
6053eb4d-56a9-415c-a3e1-6b73cb079a8a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	17043384836957645808@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 17043384836957645808			\N	f	f	f	0	\N	2024-03-03 23:59:59+00	2025-08-15 23:30:33.835994+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.835994+00	2025-08-15 23:30:33.835994+00
cf3a115c-ead1-44c2-86c4-d37ee6281715	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	18051977233055947768@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 18051977233055947768			\N	f	f	f	0	\N	2024-02-17 23:59:59+00	2025-08-15 23:30:33.864831+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.864831+00	2025-08-15 23:30:33.864831+00
5d9cadc5-c81e-4985-be45-0adac50d1463	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	6b292380-951b-46d7-8c68-888436e0b35e@au-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@au-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: 740aee6575f7007cc170759202c04ea738179e92809c979d2d043a5c0bfdc2f8			\N	f	f	f	0	\N	2024-03-02 23:48:09+00	2025-08-15 23:30:42.713551+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.713551+00	2025-08-15 23:30:42.713551+00
5b44bba4-e543-4a0e-9248-7315cf53c267	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	4a4e1bdf-9a63-459e-aa9c-061ebe07b7f9@au-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@au-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: 8f95757f6ab2d0f7b670bd5a2868034d20e7933cf117c8939d585dfdd6cfb122			\N	t	f	f	0	\N	2024-02-23 20:07:12+00	2025-08-15 23:30:33.869048+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.869048+00	2025-08-16 00:18:14.93046+00
d70bf821-e6dc-4bb8-994b-478a0c0c4be5	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAJtMojaXxgDD=OD_8Ato=vFe5wiB+6GynZONoUnX4MZ9wYZG1w@mail.gmail.com	<CAO+uLtXVeESBHQFg98M5Ukt6cBQq0T2++MG4UC9CKVTgzciRVw@mail.gmail.com>	\N	support@snorkell.ai	Allen <allen@markwm.com>	Allen Patel	{"Suman Saurabh <sumansaurabh@snorkell.ai>",saby@themarkway.co}	{"support@snorkell.ai <support@snorkell.ai>"}	\N	\N	Re: Regarding press coverage for Snorkell	Hi Suman,\r\n\r\nThanks for getting back to me, I really appreciate it.\r\n\r\nPlease feel free to schedule a call with our Account Manager, Saby using\r\nthe Calendly link below -\r\nhttps://calendly.com/saby-93k/30min\r\n\r\nIn case none of the available slots work for you please let me know some\r\ntimes that work for you in the coming week.\r\n\r\nBest,\r\nAllen\r\n\r\nOn Thu, 16 May 2024 at 17:46, Suman Saurabh <sumansaurabh@snorkell.ai>\r\nwrote:\r\n\r\n> Allen, I am interested in your service. Can we schedule a call?\r\n>\r\n> On Thu, 16 May 2024 at 1:49 PM, Allen Patel <allen@markway-abm.com> wrote:\r\n>\r\n>> Hi,\r\n>>\r\n>>\r\n>>\r\n>> I have been trying to reach out to you for the past couple of days, so\r\n>> just in case you missed my previous two emails -\r\n>>\r\n>>\r\n>>\r\n>> We are a PR agency that specializes in working with startups and getting\r\n>> them featured in some of the top publications based in the US, Canada, UK,\r\n>> and Europe such as TechCrunch, TheNextWeb, Daily Mail, Forbes,\r\n>> Entrepreneur, USA Today, and Harvard Business Review to name a few. If you\r\n>> are struggling to get your product into the limelight or need to attract\r\n>> attention from reputed VCs/investors, or simply just want to get more\r\n>> eyeballs onto your website we can help.\r\n>>\r\n>>\r\n>>\r\n>> If you are interested in learning more, just revert to me and I shall be\r\n>> happy to share more details with you.\r\n>>\r\n>>\r\n>>\r\n>>\r\n>>\r\n>>\r\n>> Allen Patel\r\n>>\r\n>> Director of Growth\r\n>>\r\n>> Marketing | MarkWay Solutions\r\n>> [image: mobilePhone] +1 615-745-9345 <+1%20615-745-9345>\r\n>> [image: emailAddress] allen@markway-abm.com\r\n>> [image: website] https://www.themarkway.com/\r\n>> [image: address] 701 Tillery Street, Unit 12 1165, Austin, Texas 78702,\r\n>> United States\r\n>> <https://www.google.com/maps/search/701+Tillery+Street,+Unit+12+1165,+Austin,+Texas+78702,+United+States?entry=gmail&source=g>\r\n>> [image: linkedin]\r\n>> <https://www.linkedin.com/company/markway-your-marketing-gateway/>\r\n>>\r\n>>\r\n>>\r\n>>\r\n>>\r\n>> If you'd like me to stop sending you emails, please click here\r\n>> <https://themarkway-co.itcn1.com/unsubscribe?email_id=MTE5MzgzNQiQLiQL&hash=bc8fbca6-3fdf-4eef-9d25-cdb19305dcfb>\r\n>>\r\n>\r\n	<div dir="ltr">Hi Suman,<div><br></div><div>Thanks for getting back to me, I really appreciate it.</div><div><br></div><div>Please feel free to schedule a call with our Account Manager, Saby using the Calendly link below -</div><div><a href="https://calendly.com/saby-93k/30min">https://calendly.com/saby-93k/30min</a><br></div><div><br></div><div>In case none of the available slots work for you please let me know some times that work for you in the coming week.</div><div><br></div><div>Best,</div><div>Allen</div></div><br><div class="gmail_quote"><div dir="ltr" class="gmail_attr">On Thu, 16 May 2024 at 17:46, Suman Saurabh &lt;<a href="mailto:sumansaurabh@snorkell.ai">sumansaurabh@snorkell.ai</a>&gt; wrote:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex"><div dir="ltr"><div dir="ltr"><div dir="auto">Allen, I am interested in your service. Can we schedule a call?</div><div></div></div><div><br><div class="gmail_quote"><div dir="ltr" class="gmail_attr">On Thu, 16 May 2024 at 1:49 PM, Allen Patel &lt;<a href="mailto:allen@markway-abm.com" target="_blank">allen@markway-abm.com</a>&gt; wrote:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex"><div><div id="m_-8103525586553605300m_8661722888691247068m_7195226512314557545email-1234567890"><span id="m_-8103525586553605300m_8661722888691247068m_7195226512314557545MTE5MzgzNQiQLiQL"></span> <p style="margin:0px"><span style="display:inline">Hi,</span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline">I have been trying to reach out to you for the past couple of days, so just in case you missed my previous two emails -</span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline">We are a PR agency that specializes in working with startups and getting them featured in some of the top publications based in the US, Canada, UK, and Europe such as TechCrunch, TheNextWeb, Daily Mail, Forbes, Entrepreneur, USA Today, and Harvard Business Review to name a few. If you are struggling to get your product into the limelight or need to attract attention from reputed VCs/investors, or simply just want to get more eyeballs onto your website we can help.</span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline">If you are interested in learning more, just revert to me and I shall be happy to share more details with you.</span></p></div></div><div><div id="m_-8103525586553605300m_8661722888691247068m_7195226512314557545email-1234567890"><p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="vertical-align:middle;font-family:Arial"><h2 color="#000000" style="margin:0px;font-size:18px;font-weight:600;font-family:Arial;color:rgb(0,0,0)"><span style="font-family:Arial">Allen</span><span style="font-family:Arial"> </span><span style="font-family:Arial">Patel</span></h2><p color="#000000" style="margin:0px;font-size:14px;line-height:22px;font-family:Arial;color:rgb(0,0,0)"><span style="font-family:Arial">Director of Growth</span></p><p color="#000000" style="margin:0px;font-weight:500;font-size:14px;line-height:22px;font-family:Arial;color:rgb(0,0,0)"><span style="font-family:Arial">Marketing</span><span style="font-family:Arial"> | </span><span style="font-family:Arial">MarkWay Solutions</span></p></td><td width="30" style="font-family:Arial"><div style="width:30px;font-family:Arial"></div></td><td color="#f86295" width="1" height="auto" style="width:1px;border-bottom:medium none currentcolor;border-left:1px solid rgb(248,98,149);font-family:Arial"></td><td width="30" style="font-family:Arial"><div style="width:30px;font-family:Arial"></div></td><td style="vertical-align:middle;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr height="25" style="vertical-align:middle;font-family:Arial"><td width="30" style="vertical-align:middle;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="vertical-align:bottom;font-family:Arial"><span color="#f86295" width="11" style="display:inline-block;font-family:Arial;background-color:rgb(248,98,149)"><img src="https://cdn2.hubspot.net/hubfs/53/tools/email-signature-generator/icons/phone-icon-2x.png" color="#f86295" alt="mobilePhone" width="13" style="display: block; font-family: Arial; background-color: rgb(248, 98, 149);"></span></td></tr></tbody></table></td><td style="padding:0px;font-family:Arial;color:rgb(0,0,0)"><a href="tel:+1%20615-745-9345" color="#000000" style="text-decoration:none;font-size:12px;font-family:Arial;color:rgb(0,0,0)" target="_blank"><span style="font-family:Arial">+1 615-745-9345</span></a></td></tr><tr height="25" style="vertical-align:middle;font-family:Arial"><td width="30" style="vertical-align:middle;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="vertical-align:bottom;font-family:Arial"><span color="#f86295" width="11" style="display:inline-block;font-family:Arial;background-color:rgb(248,98,149)"><img src="https://cdn2.hubspot.net/hubfs/53/tools/email-signature-generator/icons/email-icon-2x.png" color="#f86295" alt="emailAddress" width="13" style="display: block; font-family: Arial; background-color: rgb(248, 98, 149);"></span></td></tr></tbody></table></td><td style="padding:0px;font-family:Arial"><a href="mailto:allen@markway-abm.com" color="#000000" style="text-decoration:none;font-size:12px;font-family:Arial;color:rgb(0,0,0)" target="_blank"><span style="font-family:Arial">allen@markway-abm.com</span></a></td></tr><tr height="25" style="vertical-align:middle;font-family:Arial"><td width="30" style="vertical-align:middle;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="vertical-align:bottom;font-family:Arial"><span color="#f86295" width="11" style="display:inline-block;font-family:Arial;background-color:rgb(248,98,149)"><img src="https://cdn2.hubspot.net/hubfs/53/tools/email-signature-generator/icons/link-icon-2x.png" color="#f86295" alt="website" width="13" style="display: block; font-family: Arial; background-color: rgb(248, 98, 149);"></span></td></tr></tbody></table></td><td style="padding:0px;font-family:Arial"><a href="https://www.themarkway.com/" color="#000000" style="text-decoration:none;font-size:12px;font-family:Arial;color:rgb(0,0,0)" target="_blank"><span style="font-family:Arial">https://www.themarkway.com/</span></a></td></tr><tr height="25" style="vertical-align:middle;font-family:Arial"><td width="30" style="vertical-align:middle;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="vertical-align:bottom;font-family:Arial"><span color="#f86295" width="11" style="display:inline-block;font-family:Arial;background-color:rgb(248,98,149)"><img src="https://cdn2.hubspot.net/hubfs/53/tools/email-signature-generator/icons/address-icon-2x.png" color="#f86295" alt="address" width="13" style="display: block; font-family: Arial; background-color: rgb(248, 98, 149);"></span></td></tr></tbody></table></td><td style="padding:0px;font-family:Arial"><span color="#000000" style="font-size:12px;font-family:Arial;color:rgb(0,0,0)"><span style="font-family:Arial"><a href="https://www.google.com/maps/search/701+Tillery+Street,+Unit+12+1165,+Austin,+Texas+78702,+United+States?entry=gmail&amp;source=g" style="font-family:Arial" target="_blank">701 Tillery Street, Unit 12 1165, Austin, Texas 78702, United States</a></span></span></td></tr></tbody></table></td></tr></tbody></table></td></tr><tr style="font-family:Arial"><td style="font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="width:100%;vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td height="30" style="font-family:Arial"></td></tr><tr style="font-family:Arial"><td color="#f86295" width="auto" height="1" style="width:100%;border-bottom:1px solid rgb(248,98,149);border-left:medium none currentcolor;display:block;font-family:Arial"></td></tr><tr style="font-family:Arial"><td height="30" style="font-family:Arial"></td></tr></tbody></table></td></tr><tr style="font-family:Arial"><td style="font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="width:100%;vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="font-family:Arial"><td style="text-align:right;vertical-align:top;font-family:Arial"><table cellpadding="0" cellspacing="0" border="0" style="display:inline-block;vertical-align:-webkit-baseline-middle;font-size:medium;font-family:Arial"><tbody style="font-family:Arial"><tr style="text-align:right;font-family:Arial"><td style="font-family:Arial"><a href="https://www.linkedin.com/company/markway-your-marketing-gateway/" color="#7075db" style="display:inline-block;padding:0px;font-family:Arial;background-color:rgb(112,117,219)" target="_blank"><img src="https://cdn2.hubspot.net/hubfs/53/tools/email-signature-generator/icons/linkedin-icon-2x.png" alt="linkedin" color="#7075db" width="24" style="max-width: 135px; display: block; font-family: Arial; background-color: rgb(112, 117, 219);"></a></td><td width="5" style="font-family:Arial"><div style="font-family:Arial"></div></td></tr></tbody></table></td></tr></tbody></table></td></tr></tbody></table>\r\n<p style="margin:0px"><span style="display:inline-block">  </span></p><p style="margin:0px"><span style="display:inline-block">  </span></p><p><span>If you&#39;d like me to stop sending you emails, please <a href="https://themarkway-co.itcn1.com/unsubscribe?email_id=MTE5MzgzNQiQLiQL&amp;hash=bc8fbca6-3fdf-4eef-9d25-cdb19305dcfb" target="_blank">click here</a></span></p><img src="https://themarkway-co.itcn1.com/api/mailing/track-open/MTE5MzgzNQiQLiQL.gif?hash=bc8fbca6-3fdf-4eef-9d25-cdb19305dcfb" width="2px" height="2px"></div></div>\r\n</blockquote></div></div>\r\n<img src="https://d5fzzV04.na1.hs-salescrm-engage.com/Cto/5H+23284/d5fzzV04/R5R8b466cN61LhwV2fDmLW1Qs2WR3GHJ57W3JGLQL1W_QmlW1GB6Bc1T-N19W1N7JDW3H34JQW1N4cH21S32g2n22TGmH4W1" alt="" height="1" width="1" style="display: none;"><div></div></div>\r\n</blockquote></div>\r\n	Hi Suman, Thanks for getting back to me, I really appreciate it. Please feel free to schedule a call with our Account Manager, Saby using the Calendly...	f	t	f	0	\N	2024-05-16 23:02:27+00	2025-08-15 23:30:33.872678+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:33.872678+00	2025-08-15 23:30:33.872678+00
248126da-8fd1-4f59-9b5f-8b33146d2a55	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	7858463993074640822@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 7858463993074640822			\N	f	f	f	0	\N	2024-02-07 23:59:59+00	2025-08-15 23:30:42.767273+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.767273+00	2025-08-15 23:30:42.767273+00
a17707d1-6be4-44e5-9c03-8a6b3d46cb7b	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	56c28f2f-19c8-46e3-974c-e399e2f52fe9@au-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@au-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: 443fef2cbb2468ef8bf53b09d221fc0623d7e675787242c564c93b55e90d84bf			\N	f	f	f	0	\N	2024-02-17 18:21:27+00	2025-08-15 23:30:42.771991+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.771991+00	2025-08-15 23:30:42.771991+00
cc8d533c-875b-4faa-9e07-03ab67a83503	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	dce761f0-e301-11ee-83f6-09c26dcd2a60@facebookmail.com	\N	\N	support@snorkell.ai	Facebook <notification@facebookmail.com>	Facebook	{support@snorkell.ai}	\N	\N	\N	Confirm your business email address	Hi,\r\n\r\n \r\n\r\n   Please confirm your email address Please click the link below to confirm that your email address for snorkell.ai should be updated to support@snorkell.ai.  Confirm Now    \r\n\r\n \r\n\r\nThanks,\r\nThe Facebook team\r\n\r\n\r\n\r\n========================================\r\nThis message was sent to support@snorkell.ai at your request.\r\nMeta Platforms, Inc., Attention: Community Support, 1 Meta Way, Menlo Park, CA 94025\r\n\r\n	<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional //EN"><html><head><title>Facebook</title><meta http-equiv="Content-Type" content="text/html; charset=utf-8" /><style nonce="b0Iet2OD">@media all and (max-width: 480px){*[class].ib_t{min-width:100% !important}*[class].ib_row{display:block !important}*[class].ib_ext{display:block !important;padding:10px 0 5px 0;vertical-align:top !important;width:100% !important}*[class].ib_img,*[class].ib_mid{vertical-align:top !important}*[class].mb_blk{display:block !important;padding-bottom:10px;width:100% !important}*[class].mb_hide{display:none !important}*[class].mb_inl{display:inline !important}*[class].d_mb_flex{display:block !important}}.d_mb_show{display:none}.d_mb_flex{display:flex}@media only screen and (max-device-width: 480px){.d_mb_hide{display:none !important}.d_mb_show{display:block !important}.d_mb_flex{display:block !important}}.mb_text h1,.mb_text h2,.mb_text h3,.mb_text h4,.mb_text h5,.mb_text h6{line-height:normal}.mb_work_text h1{font-size:18px;line-height:normal;margin-top:4px}.mb_work_text h2,.mb_work_text h3{font-size:16px;line-height:normal;margin-top:4px}.mb_work_text h4,.mb_work_text h5,.mb_work_text h6{font-size:14px;line-height:normal}.mb_work_text a{color:#1270e9}.mb_work_text p{margin-top:4px}</style></head><table border="0" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td width="100%" align="center" style=""><table border="0" cellspacing="0" cellpadding="0" align="center" style="border-collapse:collapse;"><tr><td width="1204" align="center" style=""><body style="max-width:602px;margin:0 auto;" dir="ltr" bgcolor="#f6f7f9"><table border="0" cellspacing="0" cellpadding="0" align="center" id="email_table" style="border-collapse:collapse;max-width:602px;margin:0 auto;"><tr><td id="email_content" style="font-family:Helvetica Neue,Helvetica,Lucida Grande,tahoma,verdana,arial,sans-serif;background:#f6f7f9;"><table border="0" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr style=""><td height="20" style="line-height:20px;" colspan="3">&nbsp;</td></tr><tr><td height="1" colspan="3" style="line-height:1px;"><span style="color:#FFFFFF;font-size:1px;opacity:0;">      Please confirm your email address Please click the link below to confirm that your email address for snorkell.ai should be updated to support&#064;snorkell.ai.   Confirm Now        </span></td></tr><tr><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td><td style=""><table border="0" cellspacing="0" cellpadding="0" align="left" style="border-collapse:collapse;width:320px;"><tr><td style=""><tr style=""><td height="0" style="line-height:0px;">&nbsp;</td></tr></td></tr><tr><td style=""><table border="0" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style=""><td width="16" style="display:block;width:16px;">&nbsp;&nbsp;&nbsp;</td></td><td style=""><a href="https://business.facebook.com/" style="color:#1b74e4;text-decoration:none;padding:0 8px 0 0;"><img width="24" height="24" src="https://facebook.com/images/email/meta_icon.png" style="border:0;" /></a></td><td style=""><span style="font-size:20px;line-height:24px;color:#1d2129;font-family:FreightSansLFPro-Light, Helvetica Neue, Helvetica, Arial, sans-serif;">Business Manager</span></td></tr></table></td></tr><td style=""><tr style=""><td height="12" style="line-height:12px;">&nbsp;</td></tr></td><tr><td style="border-top:none;"></td></tr></table></td><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td></tr><tr><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td><td style=""><table border="0" width="100%" cellspacing="0" cellpadding="0" bgcolor="#ffffff" align="center" style="border-collapse:collapse;"><tr><td style=""><table border="0" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style="font-size:11px;font-family:LucidaGrande,tahoma,verdana,arial,sans-serif;padding-bottom:0px;"><tr style=""><td height="20" style="line-height:20px;">&nbsp;</td></tr></td></tr><tr><td style="font-size:11px;font-family:LucidaGrande,tahoma,verdana,arial,sans-serif;padding-top:0px;padding-bottom:0px;"><tr><td style=""><td width="11" style="display:block;width:11px;">&nbsp;&nbsp;&nbsp;</td></td><td style="text-align:left;"><table border="0" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style="font-size:11px;font-family:LucidaGrande,tahoma,verdana,arial,sans-serif;"><table border="0" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style="font-size:11px;font-family:LucidaGrande,tahoma,verdana,arial,sans-serif;"><span class="mb_text" style="font-family:Arial, sans-serif;font-size:12px;line-height:18px;color:#4b4f56;"><div style="font-family:Helvetica Neue,Helvetica,Arial;font-size: 12px; box-sizing: border-box; padding: 12px 20px;"><div style="font-size: 14px;"><b>Please confirm your email address</b></div><div style="margin: 12px 0px 24px 0px;"> Please click the link below to confirm that your email address for <a href="https://business.facebook.com/?business_id=721087022849264" style="text-decoration: none">snorkell.ai</a> should be updated to support&#064;snorkell.ai. </div><div><table border="0" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr style=""><td height="2" style="line-height:2px;" colspan="3">&nbsp;</td></tr><tr><td class="mb_blk" style=""><a href="https://business.facebook.com/verify/email/checkpoint/?token=Abqz2WVT5rAf-qKjmuxYb1FLv6l6vxjn7WYBZ8ZKtlhFUOSu1ZurPaDnkOo5sF4gKCAm2nc_VKKfTzmE8QxqJea_VoHc-I032-GL8PBzZ1kZ1xt1WaUZegBhlRFUaRJUgg1mfY0ds5X960sK5IdHICPYKsA9u-8TR4UEHLYxdfrZ4HAGwNvnux2lEKUyb3rETRJMKA1Kdy-QOIwyCNIDBahns6ibxHMgjMQdW2Lo85RQvIMc6qTjoMxab8CLXbpKdEo" style="color:#1b74e4;text-decoration:none;"><table border="0" width="100%" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style="border-collapse:collapse;border-radius:6px;text-align:center;display:block;background:#1877f2;background-color:#3b7cff;border:none;padding:8px 16px 10px 16px;"><a href="https://business.facebook.com/verify/email/checkpoint/?token=Abqz2WVT5rAf-qKjmuxYb1FLv6l6vxjn7WYBZ8ZKtlhFUOSu1ZurPaDnkOo5sF4gKCAm2nc_VKKfTzmE8QxqJea_VoHc-I032-GL8PBzZ1kZ1xt1WaUZegBhlRFUaRJUgg1mfY0ds5X960sK5IdHICPYKsA9u-8TR4UEHLYxdfrZ4HAGwNvnux2lEKUyb3rETRJMKA1Kdy-QOIwyCNIDBahns6ibxHMgjMQdW2Lo85RQvIMc6qTjoMxab8CLXbpKdEo" style="color:#1b74e4;text-decoration:none;display:block;"><center><font size="3"><span style="font-family:Helvetica Neue,Helvetica,Lucida Grande,tahoma,verdana,arial,sans-serif;white-space:nowrap;font-weight:bold;vertical-align:middle;color:white;font-family:Arial-BoldMT, sans-serif;text-shadow:none;white-space:nowrap;font-size:12px;line-height:14px;">Confirm&nbsp;Now</span></font></center></a></td></tr></table></a></td><td width="100%" class="mb_hide" style=""></td></tr><tr style=""><td height="0" style="line-height:0px;" colspan="3">&nbsp;</td></tr></table></div></div></span></td></tr></table></td></tr></table></td><td style=""><td width="11" style="display:block;width:11px;">&nbsp;&nbsp;&nbsp;</td></td></tr></td></tr><tr><td style="font-size:11px;font-family:LucidaGrande,tahoma,verdana,arial,sans-serif;padding-top:0px;"><tr style=""><td height="20" style="line-height:20px;">&nbsp;</td></tr></td></tr></table></td></tr></table></td><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td></tr><tr><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td><td style=""><table border="0" cellspacing="0" cellpadding="0" align="left" style="border-collapse:collapse;padding:0 0 0 15px;"><tr><td style="border-top:none;"></td></tr><tr><tr style=""><td height="15" style="line-height:15px;">&nbsp;</td></tr></tr><tr><td style=""><table border="0" cellspacing="0" cellpadding="0" style="border-collapse:collapse;"><tr><td style=""><td width="16" style="display:block;width:16px;">&nbsp;&nbsp;&nbsp;</td></td><td style=""><div style="font-size:12px;line-height:14px;font-family:Arial, sans-serif;color:#90949c;margin:0 auto 10px auto;"><a href="https://www.facebook.com/#" style="color:#90949c;text-decoration:underline;font-size:12px;line-height:14px;font-family:SFUIDisplay-BOLD, sans-serif;"></a></div><div style="font-size:12px;line-height:14px;font-family:Arial, sans-serif;color:#90949c;margin:0 auto 10px auto;">This message was sent to support&#064;snorkell.ai. Meta Platforms, Inc., Attention: Community Support, 1 Meta Way, Menlo Park, CA 94025</div></td><td style=""><td width="16" style="display:block;width:16px;">&nbsp;&nbsp;&nbsp;</td></td></tr></table></td></tr></table></td><td width="15" style="display:block;width:15px;">&nbsp;&nbsp;&nbsp;</td></tr><tr style=""><td height="20" style="line-height:20px;" colspan="3">&nbsp;</td></tr></table><span style=""><img src="https://www.facebook.com/email_open_log_pic.php?mid=HMTcxMDUyOTUyODMzODUyNDpzdXBwb3J0QHNub3JrZWxsLmFpOjg1Mw" style="border:0;width:1px;height:1px;" /></span></td></tr></table></body></td></tr></table></td></tr></table></html>\r\n\r\n\r\n	Hi,      Please confirm your email address Please click the link below to confirm that your email address for snorkell.ai should be updated to sup...	f	f	f	0	\N	2024-03-15 19:26:03+00	2025-08-15 23:31:01.945739+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.945739+00	2025-08-15 23:31:01.945739+00
2b463cf1-1707-4310-8158-b874d007b129	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	MwIPp704Q-ucgHlaRcvxtw@geopod-ismtpd-0	\N	\N	support@snorkell.ai	hey@grown.run	\N	{support@snorkell.ai}	\N	\N	\N	Want to gain more exposure at Snorkell?	Hi there,\r\n\r\nI stumbled upon Snorkell through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that automatically handles coding documentation while developers focus on writing code.\r\n\r\nI oversee a database comprising 1,700+ platforms that offer exposure to up to 13 million people monthly. Notably, clients like Formwise have rapidly achieved an MRR of $104k/month by using this very database. I'm reaching out to you because I think that if you had access to GROWN you would be able to see tremendous results.\r\n\r\nThe reason I'm letting you know is because we only have two sales periods a year, and the current one is closing on March 31st.\r\n\r\nBy using GROWN I strongly belive that you and Snorkell can expect:\r\n\r\n* Reaching possibly millions of eyes in form of exposure which leads to paying customers\r\n* High quality backlinks which results for SEO and good reputation (very important)\r\n\r\nReply if you're interested or check our website ( https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3D_lkF_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBpjaWx5FI-2FzWcBiBtypCoRVxSzeea7Z8-2BlT80ylDBDbRRhPoR-2Br3pSTA1My74zJ8dz8IC0-2FoK30gFnWMUkimYWr-2B3xA0aN-2FfWPFnbiZL1s-2BgSymjhyXSsaoFrv6B76oNQn-2BYmU7TNYvE1AeqbMXI-2BTFD5hGwPbhAWIgW-2B-2BTl9pjy ) out. Let me know if you want a sample or anything else.\r\n\r\nBest regards,\r\nSebastian Trygg, GROWN\r\nCo-founder | Chief curator\r\nwww.grown.run ( https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3DQ-4o_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBpjaWx5FI-2FzWcBiBtypCoRVxSzeea7Z8-2BlT80ylDBDbRUxQsXhSVS-2Bnwx0L-2FAP1oyXBP68eU2znw-2FwKxgYiu-2BrdJm-2FzScM-2B7A2N8kVlAz1Fp4oQYfd7fVtFs7b4KYCyqJQ5zPcgn281ZeEW8xS-2FWrzJeMT3AjsM1YlMv8P8wmUgC )	<html>\r\n  <head>\r\n    <title></title>\r\n  </head>\r\n<body>\r\n  <div data-role="module-unsubscribe" class="module" role="module" data-type="unsubscribe" style="color:#444444; font-size:12px; line-height:20px; padding:5px 6px; text-align:start;" data-muid="4e838cf3-9892-4a6d-94d6-170e474d21e5">\r\n<p>\r\nHi there,\r\n<br/><br/>\r\nI stumbled upon Snorkell through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that automatically handles coding documentation while developers focus on writing code.\r\n  <br><br>\r\nI oversee a database comprising 1,700+ platforms that offer exposure to up to 13 million people monthly. Notably, clients like Formwise have rapidly achieved an MRR of $104k/month by using this very database. I'm reaching out to you because I think that if you had access to GROWN you would be able to see tremendous results. \r\n\r\n<br><br>\r\n\r\nThe reason I'm letting you know is because we only have two sales periods a year, and the current one is closing on March 31st.\r\n<br><br>\r\nBy using GROWN I strongly belive that you and Snorkell can expect:\r\n<ul>\r\n    <li>Reaching possibly millions of eyes in form of exposure which leads to paying customers</li>\r\n    <li>High quality backlinks which results for SEO and good reputation (very important)</li>\r\n</ul>\r\n<br>\r\nReply if you're interested or check <a href="https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3DwwOu_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBpjaWx5FI-2FzWcBiBtypCoRVxSzeea7Z8-2BlT80ylDBDbRLbYJ6Pq5VagYdjD7FJq5L3Y81is5qE6-2BtvdXEoRxMrHvITWW349bdIyj5-2FAHMfSejUTUXOMPErQDK9Roe0FtcSNt1hVHpgwzi5XSF4jqcEYorQ-2FPgap1Pa2CKOsTUqvZ">our website</a> out. Let me know if you want a sample or anything else.\r\n<br><br>\r\nBest regards,<br/>Sebastian Trygg, GROWN<br/>Co-founder | Chief curator\r\n<br/>\r\n<a href="https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3D1bOu_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBpjaWx5FI-2FzWcBiBtypCoRVxSzeea7Z8-2BlT80ylDBDbRhgTLCWjWzQ4vZ-2BU6eycxwdrKJWGqW-2BONmm-2FOMOepsel0bWYA87StoeFx8fXsYzul2nyAUa6jDrQfWQDYum58zEgcMf1d7EhT86QZHEzrhi6S6jHd99ijyBBe-2B-2BPcr1RU">www.grown.run</a>\r\n\r\n</p>\r\n  </div>\r\n<img src="https://u33516475.ct.sendgrid.net/wf/open?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIj-2BQXkp53iwc-2BPfpa0BNDD-2B0trSrDasDuYJaSDHkU4LZjGR47kQx6ryRE-2FM5UA2gT97bYi8pHs4CEau0S-2BkXbs7exFGJUO0o-2FBlobEA9PRjTpzec4UNA6BjAf-2BDfhBG4yJc84VPaFhKEkUAYhDrMkcCshxlqazdw9dH0IXdzK749BdTPp6cx80xcbQQBsB-2FHsuv5czt7OgGbAdKBUSBr9cMO1YwhDjMjM42ghIXGH7W6MCZrjCDoPpQnyeib1UFZKcrfjxib8D4HCY1jV4dkIVzaB8h6qvsKabTtK7MEegxlA-3D-3D" alt="" width="1" height="1" border="0" style="height:1px !important;width:1px !important;border-width:0 !important;margin-top:0 !important;margin-bottom:0 !important;margin-right:0 !important;margin-left:0 !important;padding-top:0 !important;padding-bottom:0 !important;padding-right:0 !important;padding-left:0 !important;"/></body>\r\n</html>\r\n	Hi there, I stumbled upon Snorkell through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that automatically handles codi...	t	f	f	0	\N	2024-03-09 11:48:17+00	2025-08-15 23:30:42.752019+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:30:42.752019+00	2025-08-16 00:17:51.894898+00
b12920a6-91ff-49e0-98f6-f71d40e4a585	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	00352a7c-6afb-4d27-ac14-1734cfb33f3a@au-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@au-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: b43ead9eaf3a7a66617b3e84e53a4c252a1eed1e444532751d15d7ff3cfd1915			\N	f	f	f	0	\N	2024-03-06 20:48:38+00	2025-08-15 23:31:00.602497+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.602497+00	2025-08-15 23:31:00.602497+00
d979892d-1b49-4ee2-8daa-8c96169228d7	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	5d6A8s-lSA2HB4QnG1l89Q@geopod-ismtpd-18	\N	\N	support@snorkell.ai	hey@grown.run	\N	{support@snorkell.ai}	\N	\N	\N	Want to gain more exposure at Snorkell.ai\n\n?	Hi there,\r\n\r\nI stumbled upon Snorkell.ai\r\n\r\n through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that takes care of coding documentation while developers focus on writing code.\r\n\r\nI oversee a database comprising 1,700+ platforms that offer exposure to up to 13 million people monthly. Notably, clients like Formwise have rapidly achieved an MRR of $104k/month by using this very database. I'm reaching out to you because I think that if you had access to GROWN you would be able to see tremendous results.\r\n\r\nThe reason I'm letting you know is because we only have two sales periods a year, and the current one is closing on March 31st.\r\n\r\nBy using GROWN I strongly belive that you and Snorkell.ai\r\n\r\n can expect:\r\n\r\n* Reaching possibly millions of eyes in form of exposure which leads to paying customers\r\n* High quality backlinks which results for SEO and good reputation (very important)\r\n\r\nReply if you're interested or check our website ( https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3D4y27_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBg6ocFQaHPT-2FPTDPc1YhKFYZx4mLF38e1N3xPHwYY3D6iq7gE80KAwX0xoKRK2IlV4CnfbNKJvvxJCqkgYQg-2F-2B-2BE3tSvoFutRscl7BNNtVIqkI7iG0DecSV74pxqU5hQ8nw6ipdkipEIemWxClDWUlAF2tMmVgd60FWPXbn7Xs2j5GxiJUJy9tVGN40-2B98ePfA-3D-3D ) out. Let me know if you want a sample or anything else.\r\n\r\nBest regards,\r\nSebastian Trygg, GROWN\r\nCo-founder | Chief curator\r\nwww.grown.run ( https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3DHvXw_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBg6ocFQaHPT-2FPTDPc1YhKFYZx4mLF38e1N3xPHwYY3D6rLGrYQkMmH-2F43csrxPwlmNZIkwNXvdluogUunXhaZIo40P8tGfKoiVAqTHy1Hvhuqqjli0sgTomSkhCnVsTwLug03ncJ94nacAZzNaPgS4i316G2YrZeHUHCOPFFPkfxVQZJI4j4ky4TJUqRydcvaQ-3D-3D )	<html>\r\n  <head>\r\n    <title></title>\r\n  </head>\r\n<body>\r\n  <div data-role="module-unsubscribe" class="module" role="module" data-type="unsubscribe" style="color:#444444; font-size:12px; line-height:20px; padding:5px 6px; text-align:start;" data-muid="4e838cf3-9892-4a6d-94d6-170e474d21e5">\r\n<p>\r\nHi there,\r\n<br/><br/>\r\nI stumbled upon Snorkell.ai\r\n\r\n through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that takes care of coding documentation while developers focus on writing code.\r\n  <br><br>\r\nI oversee a database comprising 1,700+ platforms that offer exposure to up to 13 million people monthly. Notably, clients like Formwise have rapidly achieved an MRR of $104k/month by using this very database. I'm reaching out to you because I think that if you had access to GROWN you would be able to see tremendous results. \r\n\r\n<br><br>\r\n\r\nThe reason I'm letting you know is because we only have two sales periods a year, and the current one is closing on March 31st.\r\n<br><br>\r\nBy using GROWN I strongly belive that you and Snorkell.ai\r\n\r\n can expect:\r\n<ul>\r\n    <li>Reaching possibly millions of eyes in form of exposure which leads to paying customers</li>\r\n    <li>High quality backlinks which results for SEO and good reputation (very important)</li>\r\n</ul>\r\n<br>\r\nReply if you're interested or check <a href="https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3DhhHm_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBg6ocFQaHPT-2FPTDPc1YhKFYZx4mLF38e1N3xPHwYY3D6orCyHBL7SiWVCs2D43qSIK7Lve0m0cicr-2BYSSJTts6VoIHuIyFAu6wXP8aeBq7AjQ18GOVPoOE54xqJo29AeX-2BfpQux7S4jYBpweHYUd8mEXoL4d1rLeeRJMETW3Eozrq2XZs1mJrV8wMp8wfSqlEw-3D-3D">our website</a> out. Let me know if you want a sample or anything else.\r\n<br><br>\r\nBest regards,<br/>Sebastian Trygg, GROWN<br/>Co-founder | Chief curator\r\n<br/>\r\n<a href="https://u33516475.ct.sendgrid.net/ls/click?upn=u001.LRYeW46PLgNG2DG74C-2B-2Fyr8ZKPT7jDE6jqF1XgjWB7M-3DRk6j_wmc-2BViHeaSAOdjPZKiSJ-2FbYzNdN1M6ds-2FGLhSuV73kZ96fRz3TDl5-2B26hr0vlyaUWHiXvkERPlMdGS9hLQXQzuPEWwxq-2BBgeGuLLvAdma1LZ5AcURd5ZMpCYc54tYYnLSviPLJsjnxbYDKgEXdHQBg6ocFQaHPT-2FPTDPc1YhKFYZx4mLF38e1N3xPHwYY3D6wryuScNAW1i-2FY31MWXCU7QNA4pshCri-2FGMMoZYYX1mbHaaPwVqxAbHz3W7Ty-2BYTMdxNs8g3oW-2BRtSXA7O4QLRxfupl4uFYWtqmyYdeQHwGPD-2FgqLykCi9BwZvplitATYHrcq-2BpEPVn-2BlU0LaqkLHIA-3D-3D">www.grown.run</a>\r\n\r\n</p>\r\n  </div>\r\n<img src="https://u33516475.ct.sendgrid.net/wf/open?upn=u001.0aIP81AzdOWreoEpnn5mguGOhaWyysvvQzW2OitEmIj-2BQXkp53iwc-2BPfpa0BNDD-2B0trSrDasDuYJaSDHkU4LZjGR47kQx6ryRE-2FM5UA2gT97bYi8pHs4CEau0S-2BkXbs7exFGJUO0o-2FBlobEA9PRjTvsG3gF3r9lQO8pUaMRVpQFnJ6boLtLUXW8Bl14BuyaaC-2BNHwmTGjf9HgSgE8UdnXEtM958Fl0N1GIQ57SU6IfecsUb-2FIYwwHekrHA0jEmhRQ8rd1S6lPzflvtBkKF9qnp7VkbWIDAu5cTX7YeBawJHs9ms7LxL4Sv57-2FcftVe7Vp3ehWTGUxYn9dD9zoDG8sA-3D-3D" alt="" width="1" height="1" border="0" style="height:1px !important;width:1px !important;border-width:0 !important;margin-top:0 !important;margin-bottom:0 !important;margin-right:0 !important;margin-left:0 !important;padding-top:0 !important;padding-bottom:0 !important;padding-right:0 !important;padding-left:0 !important;"/></body>\r\n</html>\r\n	Hi there, I stumbled upon Snorkell.ai through Product Hunt, and I'm genuinely intrigued by the idea of an AI-powered tool that takes care of coding do...	f	f	f	0	\N	2024-03-09 09:06:04+00	2025-08-15 23:31:00.610922+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.610922+00	2025-08-15 23:31:00.610922+00
c52964fc-6af3-4d8d-ab07-b8a682b55cb1	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	65c39c73e1e13_11f6bc1550bf@jobs-bd7979bbc-lwgtw.mail	\N	\N	support@snorkell.ai	Shopify <policy@mailer.shopify.com>	Shopify	{support@snorkell.ai}	\N	\N	\N	Your refund policy is ready!	\r\n\r\nYour refund policy is ready!\r\n\r\nThanks for using Shopify’s refund policy generator. \r\nSimply follow the link below to download your file.\r\n\r\nGet your refund policy now: https://u2435018.ct.sendgrid.net/ls/click?upn=1SVposIMapPoBUvI3-2BjV2ywU63WMXIJ3DiZT-2B4upsVAk2bDAH0UJZmJaB0O5O6bGuzRUTePFBc8do0NlgKaLt-2FCsf6QFvaBuZ4bbRzBTHDN0ElEdKmxk7mgNJkSjUZfv-2B-2FBD-2Fzh-2Fj-2FlPDK7KIkqXVWDFOKK751OSQziQvfK0vbPbQFManW7BwXl8NESiXRPPEnNhH2fEMVdE6LMibXqFSXkbh9ZKY8GJ2y8I2NoknZeZe3s7O-2Bq8O5a8ft36JzwPT2qWdXqFonu4JG63cNE1rA-3D-3DZtxa_-2BHB8d5C343hfLp7ljYtulTiO1B6VPeFdbvUwm6KB5Pq1YnL9g0KxKZY4ptL1Sa7mThv4s9oOW17ks9VanzNYeQ1U4Ss-2BnBHRGMG2o8p3J-2FaDOLGqB8qNvxFht5b-2FaflAE1JaDp1YWCWftUY2n6BKMSaJ2S9V-2B9HlK4wZzaL3wy-2FoCPdamrZ7vQ7R0F8w72gSbEoj0mK3NiFIGKBnZZFy9hz65Ezqd93hX-2FmFSnB0K-2Bfw1D06m1PtW8rVcycIYcQu-2BSLZsv0FEhDFmatMT5rJkQ-3D-3D\r\n\r\n\r\n---------------------------------\r\n\r\nSelling online or in person?\r\n\r\nTry Shopify for free for 3-days, risk free: https://u2435018.ct.sendgrid.net/ls/click?upn=1SVposIMapPoBUvI3-2BjV2ywU63WMXIJ3DiZT-2B4upsVBmuWl9jKhAw30JBK11J-2FlrEjT4_-2BHB8d5C343hfLp7ljYtulTiO1B6VPeFdbvUwm6KB5Pq1YnL9g0KxKZY4ptL1Sa7mThv4s9oOW17ks9VanzNYeQ1U4Ss-2BnBHRGMG2o8p3J-2FaDOLGqB8qNvxFht5b-2FaflAPDKOjLdZ2t3ocvIY6MJzGzAHF4KRNYx5mRthrJzB7bVrNmp19iINjWc6PxRG-2FZIqRrZAzIUrrRJggrsbIaqL8y8aHtoVqee9d09HnA1om3YPWwuCXoYiIXSzWU-2FMBH3gSwLkA5-2BHW21lnK5D5468eA-3D-3D\r\n\r\n\r\n©Shopify.com | 151 O’Connor Street, Ground floor, Ottawa, ON, Canada, K2P 2L8\r\n	<!DOCTYPE html>\r\n<html>\r\n  <head>\r\n    <meta content="text/html; charset=utf-8" http-equiv="Content-Type">\r\n    <meta content="width=device-width" name="viewport">\r\n\r\n    <title></title>\r\n    <style type="text/css">\r\n      @import url(https://cdn.shopify.com/shopify-marketing_assets/builds/103.19.0/marketing_assets/builds/fonts.css);\r\n      /*////// RESET STYLES //////*/\r\n      body, #bodyTable, #bodyCell{height:100% !important; margin:0; padding:0; width:100% !important;}\r\n      table{border-collapse:collapse;}\r\n      img, a img{border:0; outline:none; text-decoration:none;}\r\n      h1, h2, h3, h4, h5, h6{margin:0; padding:0;}\r\n      p{margin: 1em 0;}\r\n      .im{color: #fff;}\r\n\r\n      /*////// CLIENT-SPECIFIC STYLES //////*/\r\n      .ReadMsgBody{width:100%;} .ExternalClass{width:100%;} /* Force Hotmail/Outlook.com to display emails at full width. */\r\n      .ExternalClass, .ExternalClass p, .ExternalClass span, .ExternalClass font, .ExternalClass td, .ExternalClass div{line-height:100%;} /* Force Hotmail/Outlook.com to display line heights normally. */\r\n      table, td{mso-table-lspace:0pt; mso-table-rspace:0pt;} /* Remove spacing between tables in Outlook 2007 and up. */\r\n      #outlook a{padding:0;} /* Force Outlook 2007 and up to provide a "view in browser" message. */\r\n      img{-ms-interpolation-mode: bicubic;} /* Force IE to smoothly render resized images. */\r\n      body, table, td, p, a, li, blockquote{-ms-text-size-adjust:100%; -webkit-text-size-adjust:100%;} /* Prevent Windows- and Webkit-based mobile platforms from changing declared text sizes. */\r\n\r\n      @media (max-width: 600px) {\r\n        table[class="content-table"],\r\n        table[class="footer-table"],\r\n        table[class="header-table"] {\r\n          width: 96% !important;\r\n          max-width: 540px !important;\r\n        }\r\n      }\r\n\r\n      @media (max-width: 480px) {\r\n        body, .body-table, p, li {\r\n          font-size: 15px !important;\r\n          line-height: 23px !important;\r\n        }\r\n\r\n        h1 {\r\n          font-size: 22px !important;\r\n          line-height: 34px !important;\r\n        }\r\n\r\n        h2 {\r\n          font-size: 18px !important;\r\n          line-height: 27px !important;\r\n        }\r\n\r\n        h3 {\r\n          font-size: 18px !important;\r\n          line-height: 27px !important;\r\n        }\r\n\r\n        h4 {\r\n          font-size: 16px !important;\r\n          line-height: 24px !important;\r\n        }\r\n        .tool-links a {\r\n          font-size: 15px !important;\r\n          line-height: 23px !important;\r\n        }\r\n        td[class="content-table__body"],\r\n        td[class="content-table__body--padding"],\r\n        td[class="content-table__hero"] {\r\n          padding: 30px 0!important;\r\n        }\r\n        .footer-table__legal p, .footer-table__links a {\r\n          font-size: 12px !important;\r\n        }\r\n        table[class="table"] td,\r\n        table[class="table table__invoice"] td {\r\n          display: block!important;\r\n        }\r\n        table[class="table"] td:first-of-type,\r\n        table[class="table table__invoice"] td:first-of-type {\r\n          width: 100%!important;\r\n        }\r\n        table[class="table"] td:last-of-type,\r\n        table[class="table table__invoice"] td:last-of-type {\r\n          width: 100%!important;\r\n          padding-bottom: 15px;\r\n        }\r\n        table[class="table table__invoice"] td:last-of-type p,\r\n        table[class="table table__invoice"] td:last-of-type h5 {\r\n          text-align: left;\r\n        }\r\n        .hidden--mobile {display: none;}\r\n      }\r\n    </style>\r\n  </head>\r\n\r\n  <body style="width: 100%; height: 100%; background: #ebeef0; color: #767676; font-family: Helvetica Neue, Helvetica, Arial; font-size: 14px; font-weight: 300; line-height: 22px;">\r\n    <table\r\n      border="0"\r\n      cellpadding="0"\r\n      cellspacing="0"\r\n      class="body-table"\r\n      style="width: 100%; height: 100%; background: #ebeef0; color: #767676; font-family: Helvetica Neue, Helvetica, Arial; font-size: 14px; font-weight: 300; line-height: 22px;">\r\n      <tr>\r\n        <td>\r\n          <table\r\n            align="center"\r\n            border="0"\r\n            cellpadding="0"\r\n            cellspacing="0"\r\n            class="content-table"\r\n            style="width: 620px; max-width: 620px; margin: 60px auto; background: white; -webkit-border-radius: 3px; -moz-border-radius: 3px; border-radius: 3px; box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);" width="620">\r\n            <tr>\r\n              <td class="header-table__logo--center" style="padding: 10px 0; text-align: center;">\r\n                <img alt="Shopify" src="https://cdn.shopify.com/s/files/1/0070/7032/files/shopify-logo2x_no_padding_left.png?7073" width="125">\r\n              </td>\r\n            </tr>\r\n            <tr>\r\n              <td class="content-table__body" style="padding: 0px;">\r\n                <!-- START CONTENT -->\r\n\r\n                <div style="padding: 0 40px; text-align: center;">\r\n  <p style="text-align: center; color: #0A0B0D; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; text-transform: uppercase; font-size: 18px; font-weight: 700; letter-spacing: -0.36px;">\r\n    Free tools\r\n  </p>\r\n\r\n  <h1 style="margin: 0 0 12px 0; padding: 0; color: #0A0B0D; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased; text-align: center; font-size: 42px; font-weight: 700; line-height: 46px; letter-spacing: -0.84px;">\r\n    Your refund policy is ready!\r\n  </h1>\r\n\r\n  <p style="font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; padding: 8px 40px 16px; color: #1F2124; font-size: 16px; line-height: 24px; font-weight: 400; text-align: center;">\r\n    Thanks for using Shopify’s refund policy generator. <br />\r\nSimply follow the link below to download your file.\r\n  </p>\r\n\r\n  <div style="text-align: center;">\r\n    <a style="font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; color: #fff; display: inline-block; text-decoration: none; padding: 14px 26px; background: #000; border-radius: 300px; font-size: 16px; font-weight: 700; line-height: 28px; letter-spacing: -0.16px;" href="https://u2435018.ct.sendgrid.net/ls/click?upn=1SVposIMapPoBUvI3-2BjV2ywU63WMXIJ3DiZT-2B4upsVAk2bDAH0UJZmJaB0O5O6bGuzRUTePFBc8do0NlgKaLt-2FCsf6QFvaBuZ4bbRzBTHDN0ElEdKmxk7mgNJkSjUZfv-2B-2FBD-2Fzh-2Fj-2FlPDK7KIkqXVWDFOKK751OSQziQvfK0vbPbQFManW7BwXl8NESiXRPPEnNhH2fEMVdE6LMibXqFSXkbh9ZKY8GJ2y8I2NoknZeZe3s7O-2Bq8O5a8ft36JzwPT2qWdXqFonu4JG63cNE1rA-3D-3DAI97_-2BHB8d5C343hfLp7ljYtulTiO1B6VPeFdbvUwm6KB5Pq1YnL9g0KxKZY4ptL1Sa7mThv4s9oOW17ks9VanzNYeQ1U4Ss-2BnBHRGMG2o8p3J-2FaDOLGqB8qNvxFht5b-2FaflAaK8C0cIcKOCTtxM1HN-2FZ2xAFrHlasXEhPzBJOR5gCYS2-2FuZa93HNpj7Ve8R08bS73l-2BM0OE5oL579nxftuWe7KjYOMUB-2BLR7T16LBNyR1c2LwTaXGYrR67g-2FHrx9LqrtYdbMnIMSGv-2B229xx13rEPA-3D-3D">Get your refund policy now</a>\r\n  </div>\r\n\r\n\r\n  <hr style="margin: 84px 0 20px; border: 0; border-top: 1px solid #000;">\r\n\r\n  <h4 style="margin: 32px 40px 24px; text-align: center; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 24px; font-weight: 700; line-height: 28px; letter-spacing: -0.48px; color: #000;">\r\n    Interested in other tools?\r\n  </h4>\r\n</div>\r\n<div class="tool-links" style="padding: 0 15px; text-align: center;">\r\n  <div style="text-align: center; width: 30%; display: inline-block; vertical-align: top; max-width: 160px;">\r\n    <img alt="Hachful - Logo Maker" src="https://cdn.shopify.com/b/shopify-brochure2-assets/4e248fbdb255574b63e8a4a67b2f2c06.png">\r\n    <a target="_blank" style="display: block; padding: 24px 0 40px; color: #000; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 18px; font-style: normal; font-weight: 700; line-height: 21px; letter-spacing: -0.36px; text-decoration: none;" href="www.shopify.com/tools/logo-maker">Hachful - Logo Maker</a>\r\n  </div>\r\n  <div style="text-align: center; width: 30%; display: inline-block; vertical-align: top; max-width: 160px;">\r\n    <img alt="Business card maker" src="https://cdn.shopify.com/b/shopify-brochure2-assets/b6ac22394c904bfa0bf41fead45e04fa.png">\r\n    <a target="_blank" style="display: block; padding: 24px 0 40px; color: #000; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 18px; font-style: normal; font-weight: 700; line-height: 21px; letter-spacing: -0.36px; text-decoration: none;" href="www.shopify.com/tools/business-card-maker">Business Card Maker</a>\r\n  </div>\r\n  <div style="text-align: center; width: 30%; display: inline-block; vertical-align: top; max-width: 160px;">\r\n    <img alt="Online Discount Calculator" src="https://cdn.shopify.com/b/shopify-brochure2-assets/c9f67b349978d7ca39799593b7a618d8.png">\r\n    <a target="_blank" style="display: block; padding: 24px 0 40px; color: #000; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 18px; font-style: normal; font-weight: 700; line-height: 21px; letter-spacing: -0.36px; text-decoration: none;" href="www.shopify.com/tools/discount-calculator">Online Discount Calculator</a>\r\n  </div>\r\n</div>\r\n\r\n<!-- GENERIC CTA -->\r\n\r\n<div style="margin-top: 24px; padding: 64px 40px; background: linear-gradient(127deg, #DBF4FF 32.36%, #C2CFFC 100%); background-image: url('https://cdn.shopify.com/b/shopify-brochure2-assets/5f7d0576539c05c707eb43fa2d733b54.png'); background-repeat: no-repeat; background-size: cover;">\r\n  <h2 style="margin: 0 0 16px 0; padding: 0; color: #0A0B0D; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased; text-align: center; font-size: 32px; font-weight: 700; line-height: 38px; letter-spacing: -0.64px;">\r\n    Selling online or in person?\r\n  </h2>\r\n  <h3 style="margin: 0 0 24px 0; padding: 0; color: #0A0B0D; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased; text-align: center; font-size: 18px; font-weight: 400; line-height: 26px; letter-spacing: -0.18px;">\r\n    Try Shopify for free for 3-days, risk free\r\n  </h3>\r\n  <p style="color: #fff; font-size: 16px; line-height: 22px; font-weight: 300; text-align: center;">\r\n    <a\r\n      class="button button--small"\r\n      href="https://u2435018.ct.sendgrid.net/ls/click?upn=1SVposIMapPoBUvI3-2BjV2ywU63WMXIJ3DiZT-2B4upsVCc6pKdUTPBKPkZeAefXy3-2Fu8tcB7RIdrizBfopKmxMKehoxx5NI5w6V61WC8kQpm3vKYfRvzVe652z5dHGdLagi-2Bft6o96U3y8d2nHSGV6-2BZPIM0Q76wsQSPi3RHBYGy0-3Dh2qw_-2BHB8d5C343hfLp7ljYtulTiO1B6VPeFdbvUwm6KB5Pq1YnL9g0KxKZY4ptL1Sa7mThv4s9oOW17ks9VanzNYeQ1U4Ss-2BnBHRGMG2o8p3J-2FaDOLGqB8qNvxFht5b-2FaflA-2B9yMhuTMHJA8DPFhXFXmtByxMJZSnul4Nvzsoi0fFMPy3SGV0pbXfK0oBnMHA0pv-2BtLIn2boeqFP5LWWXhWfmOLJVnGUhLBaCG7Ti0pKrRoNALH3fbPzdXXDdILDAn-2BsN-2F-2BiRvG7sS9PoC2n6Fwp0w-3D-3D"\r\n      style="color: white; text-decoration: none; display: inline-block; padding: 14px 26px; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 16px; font-weight: 700; -webkit-border-radius: 30px; -moz-border-radius: 30px; border-radius: 30px; background: #000;">\r\n      Start your free trial\r\n    </a>\r\n  </p>\r\n</div>\r\n\r\n<!-- FOOTER -->\r\n\r\n<div style="background: #000; padding: 40px;" class="footer-table__legal">\r\n  <p style="color: #fff; text-align: center; font-family: ShopifySans, Helvetica Neue, Helvetica, Arial, sans-serif; font-size: 14px; font-style: normal; font-weight: 400; line-height: 21px;">\r\n    &copy; <a href="https://u2435018.ct.sendgrid.net/ls/click?upn=1SVposIMapPoBUvI3-2BjV20SJMGySeBg91jZsAHDdQh5GtzYdjBbbL87BfZLrFzCLNNMyS87JRETjjl9zhkdu8uCiMK-2BF-2BP1Q0XsjgYoqJXH0nFhF2nwn4FfW2pC-2BnwyaDpHBfZ-2FzTGGIeHzjuVmr4Q-3D-3D_Hb2_-2BHB8d5C343hfLp7ljYtulTiO1B6VPeFdbvUwm6KB5Pq1YnL9g0KxKZY4ptL1Sa7mThv4s9oOW17ks9VanzNYeQ1U4Ss-2BnBHRGMG2o8p3J-2FaDOLGqB8qNvxFht5b-2FaflA0rq00I5SNhRLeSZCoL4vtQ00s7FDh5nEJLXFEhmCLaCVU3qcZu8brV78QxifFEMwjjE3I-2F1UFxUto1ZzTgW4t-2FPeUwrnfBxOE0UnQPJ4CAt2FZHjhIpV-2B17vxDC-2F6FBcDpJuV4Iu7I5-2BMQFVZMOO9g-3D-3D" style="color: #fff; text-decoration: none;">Shopify</a>,\r\n              151 O’Connor Street, Ground floor,<br> Ottawa, ON, Canada, K2P 2L8\r\n  </p>\r\n</div>\r\n\r\n\r\n                <!-- END CONTENT -->\r\n              </td>\r\n            </tr>\r\n          </table>\r\n        </td>\r\n      </tr>\r\n    </table>\r\n  <img src="https://u2435018.ct.sendgrid.net/wf/open?upn=r3XecG9Oeir8G6iSrKDq5LaYG7HFDMRvVJ0BJDMo98LgCGimWZkP4rRUKzg1VNdlDugz1nffw1vul2529FG7aFQZoQFX2nN5JXE1uyCVrA5ejdhB3N8moJIyhjVlGBO10RH8e-2BO7-2FIuHYD05S2DWj6DixfXcVbMwqBYKszE-2BXCiNxM-2B1RPgYsyb-2B3sB23djcfjYaHfHOb06PSSyU6oTP1RPJURVthjmqgMJpML-2FEEp45Hbm3QCvcMwiG1M5GPlMQrfYrAt-2Blx45B5OeeYtIpIF1ZFZCqeFXYXacrNVW-2BDJQ-3D" alt="" width="1" height="1" border="0" style="height:1px !important;width:1px !important;border-width:0 !important;margin-top:0 !important;margin-bottom:0 !important;margin-right:0 !important;margin-left:0 !important;padding-top:0 !important;padding-bottom:0 !important;padding-right:0 !important;padding-left:0 !important;"/></body>\r\n</html>\r\n	Your refund policy is ready! Thanks for using Shopify’s refund policy generator. Simply follow the link below to download your file. Get your refund...	f	f	f	0	\N	2024-02-07 15:06:28+00	2025-08-15 23:31:00.625689+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.625689+00	2025-08-15 23:31:00.625689+00
b705df28-c528-4fbd-a3da-a438a0fe8ea2	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	6824565863066703732@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 6824565863066703732			\N	f	f	f	0	\N	2024-02-19 23:59:59+00	2025-08-15 23:31:00.632872+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.632872+00	2025-08-15 23:31:00.632872+00
92844f4c-1208-432d-b829-8b6e3dd73b31	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	12935147600994161843@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 12935147600994161843			\N	f	f	f	0	\N	2024-02-21 23:59:59+00	2025-08-15 23:31:00.637847+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.637847+00	2025-08-15 23:31:00.637847+00
0e214197-1877-45bf-aa77-6fe31e5978cd	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	14803769078276428600@google.com	\N	\N	support@snorkell.ai	noreply-dmarc-support@google.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: google.com Report-ID: 14803769078276428600			\N	f	f	f	0	\N	2024-02-15 23:59:59+00	2025-08-15 23:31:01.911948+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.911948+00	2025-08-15 23:31:01.911948+00
ad6238dc-2a1f-4346-aa22-4b7c1e863f47	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	b7295190-eb54-4114-95ba-711fdb7ba4b8@au-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@au-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: 8f95757f6ab2d0f7b670bd5a2868034d20e7933cf117c8939d585dfdd6cfb122			\N	f	f	f	0	\N	2024-02-16 18:50:02+00	2025-08-15 23:31:01.933687+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.933687+00	2025-08-15 23:31:01.933687+00
554b823f-5eca-4dcd-b746-9f2d8228e171	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	03fd735f-0513-4cc4-91fe-e98714c01b8b@us-1.mimecastreport.com	\N	\N	support@snorkell.ai	no-reply@us-1.mimecastreport.com	\N	{support@snorkell.ai}	\N	\N	\N	Report domain: snorkell.ai Submitter: mimecast.org Report-ID: a088830c2a24cf887f3a7f4d3eaf0e30f6d8407f8e0841dd9ab39797b6f3158b			\N	f	f	f	0	\N	2024-02-24 07:48:27+00	2025-08-15 23:31:01.93982+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.93982+00	2025-08-15 23:31:01.93982+00
2b807f64-4252-4aea-bfb7-b792a6d7e76a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAOne=D6w+20AT+36tkRdhQEQV9rLPESh5_V+aox7MbkOFD+8QQ@mail.gmail.com	\N	\N	support@snorkell.ai	Suman Saurabh <sumanrocs@gmail.com>	Suman Saurabh	{support@snorkell.ai}	\N	\N	\N	Test support	Test support\r\n	<div dir="ltr">Test support</div>\r\n	Test support	t	f	f	0	\N	2025-08-15 06:06:30+00	2025-08-15 23:31:00.620546+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:00.620546+00	2025-08-15 23:31:13.752114+00
cb9b9057-67ec-4135-ac0c-af758fd2c563	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAOne=D5sP+y7AOaeXASjcU0g6U5wa4efLW8=ppX37qVirj+6rw@mail.gmail.com	\N	\N	support@snorkell.ai	Suman Saurabh <sumanrocs@gmail.com>	Suman Saurabh	{support@snorkell.ai}	\N	\N	\N	test	\r\n	<div dir="ltr"><br></div>\r\n	\N	t	f	f	0	\N	2025-08-12 23:28:41+00	2025-08-15 23:31:01.964023+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.964023+00	2025-08-15 23:53:24.596384+00
cf24c681-2020-4e6a-92d9-0658ebe75a12	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	CAAxuJtEw5bcbku=-Nd-mnN5Qe_zomYgM1mSJt79jHOPCSwVJwQ@mail.gmail.com	\N	\N	support@snorkell.ai	Dee - Founder <dee@pearllemongroup.uk>	Dee - Founder	{support@snorkell.ai}	\N	\N	\N	Exciting Testimonial Offer from Pearl Lemon for Snorkell	 Exciting Testimonial Offer from Pearl Lemon for Snorkell\r\n\r\nHey team Snorkell,\r\n\r\nDee here, founder of Pearl Lemon Group.\r\n\r\nA few of our team members have recently started using Snorkell as part of\r\ntheir workflow, and they’re loving it! :)\r\n\r\nJust wanted to say—awesome job! We know how important testimonials and case\r\nstudies can be, especially when you’re building something great.\r\n\r\nAs fellow founders, we’d love to support you. If there’s any way we can\r\nhelp out with that, just give us a shout!\r\n\r\nThanks a ton!\r\n\r\n\r\n\r\n\r\n\r\n\r\n\r\n*DeeDodgy marathon runner, balding 38-year-old with far too many\r\nregrettable tattoos :)Unsubscribe\r\n<http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/3928a83c-91b6-48e4-906c-496c60707e61>*\r\n	\r\n      <html>\r\n      <head>\r\n        <title>Exciting Testimonial Offer from Pearl Lemon for Snorkell</title>\r\n        <meta content="text/html;" charset="utf-8" http-equiv="Content-Type">\r\n      </head>\r\n      <body><p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Hey team Snorkell,</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee here, founder of Pearl Lemon Group.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">A few of our team members have recently started using Snorkell as part of their workflow, and they’re loving it! :)</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Just wanted to say—awesome job! We know how important testimonials and case studies can be, especially when you’re building something great.</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">As fellow founders, we’d love to support you. If there’s any way we can help out with that, just give us a shout!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:12pt;margin-bottom:12pt"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Thanks a ton!</span></p>\r\n<p dir="ltr" style="line-height:1.38;margin-top:0pt;margin-bottom:0pt"><strong id="docs-internal-guid-ed630042-7fff-5576-cc31-593961249aed" style="font-weight:normal"><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dee<br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap"><br></span><span style="font-size:12pt;font-family:Verdana,sans-serif;color:#000000;background-color:transparent;font-weight:400;font-style:normal;font-variant:normal;text-decoration:none;vertical-align:baseline;white-space:pre-wrap">Dodgy marathon runner, balding 38-year-old with far too many regrettable tattoos :)<br><br><br><a style="color:#999;font-weight:normal;font-style:italic" href="http://w1.mssusw.com/prod/unsubscribe-confirm/2fddd33c-d578-4d7c-8506-e21232c3/support%40snorkell.ai/3928a83c-91b6-48e4-906c-496c60707e61">Unsubscribe</a><br></span></strong></p>\r\n<strong id="docs-internal-guid-5c91a878-7fff-081f-784b-47622bbf035f" style="font-weight:normal"></strong><img alt="" width="1" height="1" class="beacon-o" src="http://w1.mssusw.com/prod/open/3928a83c-91b6-48e4-906c-496c60707e61" style="float:left;margin-left:-1px;position:absolute;"></body>\r\n      </html>\r\n	Exciting Testimonial Offer from Pearl Lemon for Snorkell Hey team Snorkell, Dee here, founder of Pearl Lemon Group. A few of our team members have rec...	t	f	f	0	\N	2024-11-19 16:02:59+00	2025-08-15 23:31:01.959224+00	synced	\N	\N	f	faa8850b-44d4-4258-a45c-abde43abe35f	{}	\\x	2025-08-15 23:31:01.959224+00	2025-08-18 07:56:44.529904+00
\.


--
-- Data for Name: email_mailboxes; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_mailboxes (id, tenant_id, address, inbound_connector_id, routing_rules, allow_new_ticket, created_at, updated_at, project_id, display_name) FROM stdin;
5439ce42-40ee-4c4b-b836-2ca448467000	550e8400-e29b-41d4-a716-446655440000	support@snorkell.ai	faa8850b-44d4-4258-a45c-abde43abe35f	{}	t	2025-08-14 11:58:48.571237+00	2025-08-14 11:58:48.571237+00	550e8400-e29b-41d4-a716-446655440001	\N
\.


--
-- Data for Name: email_sync_status; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_sync_status (id, tenant_id, connector_id, mailbox_address, last_sync_at, last_uid, last_message_date, sync_status, sync_error, emails_synced_count, created_at, updated_at) FROM stdin;
7bde4d03-4a56-4339-bd1d-76e2d60025a9	550e8400-e29b-41d4-a716-446655440000	faa8850b-44d4-4258-a45c-abde43abe35f	support@snorkell.ai	2025-08-15 23:31:02.891263+00	113	2024-12-10 17:06:30+00	idle	\N	81	2025-08-15 02:05:25.127625+00	2025-08-15 23:31:02.891263+00
\.


--
-- Data for Name: email_transports; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.email_transports (id, tenant_id, outbound_connector_id, envelope_from_domain, dkim_selector, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: file_attachments; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.file_attachments (id, tenant_id, project_id, filename, original_filename, content_type, file_size, storage_provider, storage_path, storage_bucket, attachment_type, related_entity_type, related_entity_id, checksum, is_public, expires_at, uploaded_by_agent_id, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: goose_db_version; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.goose_db_version (id, version_id, is_applied, tstamp) FROM stdin;
1	0	t	2025-08-09 23:02:14.183484
2	1	t	2025-08-09 23:02:14.198928
3	2	t	2025-08-09 23:02:14.292542
4	3	t	2025-08-09 23:02:14.306215
5	4	t	2025-08-09 23:02:14.32533
6	5	t	2025-08-09 23:15:21.922109
7	6	t	2025-08-09 23:28:39.654269
8	7	t	2025-08-10 10:25:34.413557
9	8	t	2025-08-10 10:25:53.839135
10	10	t	2025-08-10 16:50:58.390111
11	11	t	2025-08-10 16:53:00.840079
12	12	t	2025-08-10 16:53:35.042281
\.


--
-- Data for Name: integration_categories; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.integration_categories (id, name, display_name, description, icon, sort_order, is_active, created_at) FROM stdin;
272c9e59-1ebd-4b43-b27d-da9655434250	communication	Communication	Chat platforms, messaging, and team collaboration tools	message-circle	10	t	2025-08-12 05:18:59.148436+00
03da4c1f-3c65-4bc4-b7aa-9a18901c3369	project_management	Project Management	Task tracking, project planning, and workflow tools	folder	20	t	2025-08-12 05:18:59.148436+00
69641fcc-c267-46de-a199-b09141704f12	crm_sales	CRM & Sales	Customer relationship management and sales tools	users	30	t	2025-08-12 05:18:59.148436+00
ddb6bf0c-4ef0-4205-982f-317e91c6b61b	support_helpdesk	Support & Helpdesk	Customer support and help desk platforms	headphones	40	t	2025-08-12 05:18:59.148436+00
8acccd19-0ce4-4365-a3b8-f8dc66c98a79	calendar_scheduling	Calendar & Scheduling	Calendar integration and appointment scheduling	calendar	50	t	2025-08-12 05:18:59.148436+00
652c747f-6574-4ea4-b03d-b825f3418ef1	file_storage	File Storage	Cloud storage and file sharing services	cloud	60	t	2025-08-12 05:18:59.148436+00
e0e06432-9378-401e-a399-f21fdb12d368	payment_billing	Payment & Billing	Payment processing and billing systems	credit-card	70	t	2025-08-12 05:18:59.148436+00
52a79cd4-2852-4388-a72e-13166711fd3d	email_marketing	Email & Marketing	Email marketing and automation platforms	mail	80	t	2025-08-12 05:18:59.148436+00
1ac1f58c-eae3-4c3c-9f96-3d51b5a050e3	development	Development	Code repositories, development tools, and DevOps	code	90	t	2025-08-12 05:18:59.148436+00
239b5236-bba6-4bf1-9820-7df43cd40207	ecommerce	E-commerce	Online stores and e-commerce platforms	shopping-cart	100	t	2025-08-12 05:18:59.148436+00
12fcd38c-7e4e-4e16-a261-f9358087c7ed	automation	Automation	Workflow automation and integration platforms	zap	110	t	2025-08-12 05:18:59.148436+00
a326a004-caed-41e2-9e68-a3756172deac	analytics	Analytics	Analytics and reporting tools	bar-chart	120	t	2025-08-12 05:18:59.148436+00
fff46033-af23-4f2c-b7d7-874f0a564e1b	social_media	Social Media	Social media platforms and management tools	share-2	130	t	2025-08-12 05:18:59.148436+00
bf63db84-831f-4437-87f6-40f4a0a5a6bc	custom	Custom & Webhooks	Custom integrations and webhook endpoints	settings	140	t	2025-08-12 05:18:59.148436+00
\.


--
-- Data for Name: integration_sync_logs; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.integration_sync_logs (id, tenant_id, project_id, integration_id, operation, status, external_id, request_payload, response_payload, error_message, duration_ms, created_at) FROM stdin;
\.


--
-- Data for Name: integration_templates; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.integration_templates (id, category_id, type, name, display_name, description, logo_url, website_url, documentation_url, auth_method, config_schema, default_config, supported_events, is_featured, is_active, sort_order, created_at, updated_at) FROM stdin;
103fdd30-7434-44f1-916c-fa0ecf6536e6	272c9e59-1ebd-4b43-b27d-da9655434250	slack	slack	Slack	Send notifications and manage tickets through Slack channels	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated,message.created}	t	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
d5ed6a4c-42fb-4ae4-afa5-5ef87c0419b8	272c9e59-1ebd-4b43-b27d-da9655434250	microsoft_teams	microsoft_teams	Microsoft Teams	Integrate with Microsoft Teams for notifications and collaboration	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated,message.created}	t	t	20	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
1b16131f-4d69-4fe6-8b79-56619efd9273	272c9e59-1ebd-4b43-b27d-da9655434250	discord	discord	Discord	Send notifications to Discord channels	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	30	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
38854fe7-5b5f-47fb-a917-1b350b1b8d0b	03da4c1f-3c65-4bc4-b7aa-9a18901c3369	jira	jira	Jira	Sync tickets with Jira issues for project tracking	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated,ticket.status_changed}	t	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
26882b7f-6a57-411b-8910-4182ae5ec859	03da4c1f-3c65-4bc4-b7aa-9a18901c3369	linear	linear	Linear	Create and sync issues with Linear project management	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	t	t	20	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
6516e6df-1841-417f-a48b-e5bfebfd03f3	03da4c1f-3c65-4bc4-b7aa-9a18901c3369	asana	asana	Asana	Track tickets as tasks in Asana projects	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	30	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
6ae0c66b-8b0f-4d40-9e28-93fbbeefbd45	03da4c1f-3c65-4bc4-b7aa-9a18901c3369	trello	trello	Trello	Create Trello cards from support tickets	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	40	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
f071234d-fd4f-4a07-898f-2aafb206c64d	03da4c1f-3c65-4bc4-b7aa-9a18901c3369	notion	notion	Notion	Log tickets and updates in Notion databases	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	50	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
5fa631e7-5789-4b1d-a871-6a890bbf36f7	8acccd19-0ce4-4365-a3b8-f8dc66c98a79	calendly	calendly	Calendly	Create tickets from Calendly appointments and meetings	\N	\N	\N	oauth	{}	{}	{ticket.created}	t	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
5116e7ab-fdd1-4ba0-9e6e-a3127b65cabf	8acccd19-0ce4-4365-a3b8-f8dc66c98a79	google_calendar	google_calendar	Google Calendar	Schedule follow-ups and meetings directly from tickets	\N	\N	\N	oauth	{}	{}	{ticket.updated}	f	t	20	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
3f13ffd4-6fcf-4eec-9a6d-0ef219e932b0	8acccd19-0ce4-4365-a3b8-f8dc66c98a79	outlook_calendar	outlook_calendar	Outlook Calendar	Integrate with Outlook for meeting scheduling	\N	\N	\N	oauth	{}	{}	{ticket.updated}	f	t	30	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
1a2ef8e9-ec05-4ece-8fe8-a9089f39ae7f	12fcd38c-7e4e-4e16-a261-f9358087c7ed	zapier	zapier	Zapier	Connect with 1000+ apps through Zapier automation	\N	\N	\N	api_key	{}	{}	{ticket.created,ticket.updated,ticket.status_changed,message.created}	t	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
eae96585-77cb-42de-ae9d-4dc513791d76	12fcd38c-7e4e-4e16-a261-f9358087c7ed	webhook	webhook	Custom Webhooks	Send data to any external service via HTTP webhooks	\N	\N	\N	api_key	{}	{}	{ticket.created,ticket.updated,ticket.status_changed,message.created,agent.assigned}	f	t	20	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
88edf7d0-fb6b-4615-9d2c-073cd9c82d52	1ac1f58c-eae3-4c3c-9f96-3d51b5a050e3	github	github	GitHub	Link tickets to GitHub issues and pull requests	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
dc13f6b6-b38c-46cb-9057-28e9e08fbe83	ddb6bf0c-4ef0-4205-982f-317e91c6b61b	zendesk	zendesk	Zendesk	Migrate or sync tickets with Zendesk	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	10	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
b7121b1a-a5f8-4fec-ba4c-0435415ecb11	ddb6bf0c-4ef0-4205-982f-317e91c6b61b	freshdesk	freshdesk	Freshdesk	Integrate with Freshdesk for ticket synchronization	\N	\N	\N	oauth	{}	{}	{ticket.created,ticket.updated}	f	t	20	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
d0b9624e-2c31-4e1f-94f2-dac80100bd02	ddb6bf0c-4ef0-4205-982f-317e91c6b61b	intercom	intercom	Intercom	Sync conversations and customer data with Intercom	\N	\N	\N	oauth	{}	{}	{ticket.created,message.created}	f	t	30	2025-08-12 05:20:10.621621+00	2025-08-12 05:20:10.621621+00
\.


--
-- Data for Name: integrations; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.integrations (id, tenant_id, project_id, type, name, status, config, oauth_token_id, webhook_url, webhook_secret, last_sync_at, last_error, retry_count, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: notification_deliveries; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.notification_deliveries (id, notification_id, channel, status, external_id, error_message, delivered_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.notifications (id, tenant_id, project_id, agent_id, type, title, message, priority, channels, action_url, metadata, is_read, read_at, expires_at, created_at, updated_at) FROM stdin;
15bcd137-4ab5-4ee1-967e-6bc4d1d64a6a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440030	message_received	New message from Visitor	Hello	normal	{web}	/chat/session/7ac463d5-20fe-4cf4-9da6-5623fa579c3f	{}	t	2025-08-23 01:08:28.03728+00	\N	2025-08-23 00:37:02.938325+00	2025-08-23 01:08:28.03728+00
\.


--
-- Data for Name: oauth_tokens; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.oauth_tokens (id, tenant_id, provider, account_email, access_token_enc, refresh_token_enc, expires_at, scopes, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: organizations; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.organizations (id, tenant_id, name, external_ref, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440010	550e8400-e29b-41d4-a716-446655440000	Example Corp	\N	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440011	550e8400-e29b-41d4-a716-446655440000	Test Industries	\N	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
\.


--
-- Data for Name: projects; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.projects (id, tenant_id, key, name, status, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440000	SUPPORT	Customer Support	active	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440002	550e8400-e29b-41d4-a716-446655440000	OPS	Operations	active	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
123e4567-e89b-12d3-a456-426614174001	123e4567-e89b-12d3-a456-426614174000	TEST	Test Project	active	2025-08-15 07:09:23.745931+00	2025-08-15 07:09:23.745931+00
\.


--
-- Data for Name: rate_limit_buckets; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.rate_limit_buckets (id, tenant_id, identifier, bucket_type, current_count, max_count, window_start, window_duration, last_refill, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: role_permissions; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.role_permissions (role, perm) FROM stdin;
tenant_admin	tenant.manage
tenant_admin	project.create
tenant_admin	project.manage
tenant_admin	ticket.read
tenant_admin	ticket.write
tenant_admin	ticket.assign
tenant_admin	ticket.close
tenant_admin	note.private.read
tenant_admin	note.private.write
tenant_admin	agent.manage
tenant_admin	sla.manage
tenant_admin	webhook.manage
project_admin	project.manage
project_admin	ticket.read
project_admin	ticket.write
project_admin	ticket.assign
project_admin	ticket.close
project_admin	note.private.read
project_admin	note.private.write
project_admin	agent.manage
project_admin	sla.manage
supervisor	ticket.read
supervisor	ticket.write
supervisor	ticket.assign
supervisor	ticket.close
supervisor	note.private.read
supervisor	note.private.write
agent	ticket.read
agent	ticket.write
agent	ticket.assign.self
read_only	ticket.read
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.roles (role, description) FROM stdin;
tenant_admin	Full access to tenant and all projects
project_admin	Full access to a specific project
supervisor	Read/write tickets, manage assignments in project
agent	Read/write assigned tickets in project
read_only	Read-only access to tickets in project
\.


--
-- Data for Name: schema_migrations; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.schema_migrations (version, applied_at) FROM stdin;
001_initial_schema	2025-08-12 11:17:13.292869
002_enable_rls	2025-08-12 11:17:13.292869
003_seed_data	2025-08-12 11:17:13.292869
004_email_subsystem	2025-08-12 11:18:16.957374
005_integrations_subsystem	2025-08-12 11:18:16.957374
006_advanced_features_subsystem	2025-08-12 11:18:16.957374
007_add_customer_name_to_tickets	2025-08-12 11:18:16.957374
008_api_keys_table	2025-08-12 11:18:16.957374
010_convert_roles_to_enum	2025-08-12 11:18:16.957374
011_tenant_settings	2025-08-12 11:18:16.957374
012_enhanced_integrations	2025-08-12 11:18:16.957374
013_email_inbox_system	2025-08-12 11:18:35.413213
014_project_scoped_email_system	2025-08-12 15:26:56.924463
015_unique_from_address_constraint	2025-08-14 03:32:39.228872
016_add_display_name_to_mailboxes	2025-08-14 03:33:46.973926
017_remove_email_address_columns_from_connectors	2025-08-16 09:43:06.912932
018_chat_system	2025-08-16 09:43:06.976143
019_enhanced_chat_widgets	2025-08-21 12:53:02.456934
020_ai_status_widget	2025-08-21 12:54:11.844418
021_chat_widget_background_color	2025-08-22 01:38:58.858575
\.


--
-- Data for Name: tenant_project_settings; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.tenant_project_settings (id, tenant_id, project_id, setting_key, setting_value, created_at, updated_at) FROM stdin;
df9bbb92-6ef4-4f34-acc2-7c6719a49d1f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	email_settings	{"from_name": "Sumanrocs Support", "smtp_host": "smtp.gmail.com", "smtp_port": 587, "from_email": "sumanrocs+support@gmail.com", "smtp_password": "goldenRover21a", "smtp_username": "sumanrocs@google.com", "smtp_encryption": "", "enable_email_to_ticket": true, "enable_email_notifications": true}	2025-08-11 06:40:44.233012+00	2025-08-11 06:40:44.233012+00
b3e9ffbb-fe3a-443b-ab7c-56abd21a3e9a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	branding_settings	{"logo_url": "https://www.penify.dev/penify-logo.svg", "custom_css": "", "favicon_url": "", "support_url": "https://tms.com", "accent_color": "#000000", "company_name": "Suman Test", "primary_color": "#970c0c", "secondary_color": "", "header_logo_height": 0, "enable_custom_branding": false}	2025-08-12 01:19:58.931542+00	2025-08-12 01:19:58.931542+00
\.


--
-- Data for Name: tenants; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.tenants (id, name, status, region, kms_key_id, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440000	Acme Corporation	active	us-east-1	\N	2025-08-09 23:02:14.306215+00	2025-08-09 23:02:14.306215+00
123e4567-e89b-12d3-a456-426614174000	Test Company	active	\N	\N	2025-08-15 07:09:23.743351+00	2025-08-15 07:09:23.743351+00
\.


--
-- Data for Name: ticket_mail_routing; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.ticket_mail_routing (id, tenant_id, project_id, ticket_id, public_token, reply_address, message_id_root, created_at, revoked_at) FROM stdin;
\.


--
-- Data for Name: ticket_messages; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.ticket_messages (id, tenant_id, project_id, ticket_id, author_type, author_id, body, is_private, created_at) FROM stdin;
23f2441c-415f-4302-ad69-6f33d6601108	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	Hello	f	2025-08-24 02:35:26.615797+00
9f462f80-86bb-4d03-8ff0-4ca961739f8f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	hello2	f	2025-08-24 02:35:50.89918+00
c8c16a0a-eade-453e-af8f-78278441fc76	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	wsbsbsb	f	2025-08-24 02:37:11.245567+00
ce98a6ca-3324-4735-bcc5-0167cafecefa	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	bdbfbfd	f	2025-08-24 02:38:30.020765+00
f056779c-5640-492c-b8bc-0cf073c78f0f	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	ffdfdf	f	2025-08-24 02:39:46.453953+00
92fa7b96-ed07-442b-a7ea-1c4ed1942cd6	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	rab	f	2025-08-24 02:39:59.895555+00
39350bcf-44e7-4ed9-875c-59047e3ab78e	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	gsddbss	f	2025-08-24 02:42:57.362796+00
03061efe-15f9-4341-b415-de5831879d8d	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440053	agent	550e8400-e29b-41d4-a716-446655440030	dfbdfsb	f	2025-08-24 02:44:13.003511+00
03cb59b8-a4ee-438c-b409-311a8efb11d9	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440053	agent	550e8400-e29b-41d4-a716-446655440030	sdvsdvds	f	2025-08-24 02:46:19.474947+00
306268f5-38e6-4750-8fde-d7a6407502d7	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440053	agent	550e8400-e29b-41d4-a716-446655440030	dvdvd	f	2025-08-24 02:49:19.976206+00
b2a26e8d-56b4-4ae1-8da3-1f4fc4f0f58a	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	heeello	f	2025-08-24 08:51:45.187863+00
a0ef7095-952a-492f-9604-e654be1fc6e2	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	heelllo	f	2025-08-24 08:52:04.799823+00
ad03813a-ad98-42e5-8d88-8f2679b3eab8	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	heeelo	f	2025-08-24 08:52:08.970809+00
79f13153-c8a8-4535-b501-5284e1e97c54	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	550e8400-e29b-41d4-a716-446655440050	agent	550e8400-e29b-41d4-a716-446655440030	Heeeeelllllll	f	2025-08-24 08:52:34.230766+00
\.


--
-- Data for Name: ticket_tags; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.ticket_tags (ticket_id, tenant_id, project_id, tag, created_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440050	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	mobile	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440050	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	authentication	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440051	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	feature-request	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440051	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	ui	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440052	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	performance	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440052	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	api	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440052	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	urgent	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440053	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	password	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440054	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	billing	2025-08-09 23:02:14.306215+00
550e8400-e29b-41d4-a716-446655440054	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	upgrade	2025-08-09 23:02:14.306215+00
\.


--
-- Data for Name: tickets; Type: TABLE DATA; Schema: public; Owner: tms
--

COPY public.tickets (id, tenant_id, project_id, number, subject, status, priority, type, source, customer_id, assignee_agent_id, created_at, updated_at) FROM stdin;
550e8400-e29b-41d4-a716-446655440050	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	1	Login issues with mobile app	open	high	problem	email	550e8400-e29b-41d4-a716-446655440020	550e8400-e29b-41d4-a716-446655440031	2025-08-09 23:02:14.306215+00	2025-08-10 10:25:34.413557+00
550e8400-e29b-41d4-a716-446655440051	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	2	Feature request: Dark mode	open	low	task	web	550e8400-e29b-41d4-a716-446655440021	\N	2025-08-09 23:02:14.306215+00	2025-08-10 10:25:34.413557+00
550e8400-e29b-41d4-a716-446655440052	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440002	3	Server performance issues	open	urgent	incident	api	550e8400-e29b-41d4-a716-446655440022	550e8400-e29b-41d4-a716-446655440032	2025-08-09 23:02:14.306215+00	2025-08-10 10:25:34.413557+00
550e8400-e29b-41d4-a716-446655440053	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	4	How to reset password?	resolved	normal	question	email	550e8400-e29b-41d4-a716-446655440020	550e8400-e29b-41d4-a716-446655440031	2025-08-09 23:02:14.306215+00	2025-08-10 10:25:34.413557+00
550e8400-e29b-41d4-a716-446655440054	550e8400-e29b-41d4-a716-446655440000	550e8400-e29b-41d4-a716-446655440001	5	Billing inquiry about upgrade	pending	normal	question	web	550e8400-e29b-41d4-a716-446655440021	550e8400-e29b-41d4-a716-446655440031	2025-08-09 23:02:14.306215+00	2025-08-10 10:25:34.413557+00
123e4567-e89b-12d3-a456-426614174002	123e4567-e89b-12d3-a456-426614174000	123e4567-e89b-12d3-a456-426614174001	1	Test Support Ticket	open	normal	question	web	123e4567-e89b-12d3-a456-426614174003	\N	2025-08-15 07:10:26.140792+00	2025-08-15 07:10:26.140792+00
\.


--
-- Name: goose_db_version_id_seq; Type: SEQUENCE SET; Schema: public; Owner: tms
--

SELECT pg_catalog.setval('public.goose_db_version_id_seq', 12, true);


--
-- Name: ticket_number_seq; Type: SEQUENCE SET; Schema: public; Owner: tms
--

SELECT pg_catalog.setval('public.ticket_number_seq', 1, false);


--
-- Name: agent_project_roles agent_project_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agent_project_roles
    ADD CONSTRAINT agent_project_roles_pkey PRIMARY KEY (agent_id, tenant_id, project_id);


--
-- Name: agents agents_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_pkey PRIMARY KEY (id);


--
-- Name: agents agents_tenant_id_email_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_tenant_id_email_key UNIQUE (tenant_id, email);


--
-- Name: api_keys api_keys_key_hash_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: attachments attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_pkey PRIMARY KEY (id);


--
-- Name: chat_messages chat_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_pkey PRIMARY KEY (id);


--
-- Name: chat_session_participants chat_session_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_session_participants
    ADD CONSTRAINT chat_session_participants_pkey PRIMARY KEY (session_id, agent_id);


--
-- Name: chat_sessions chat_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_pkey PRIMARY KEY (id);


--
-- Name: chat_widgets chat_widgets_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_widgets
    ADD CONSTRAINT chat_widgets_pkey PRIMARY KEY (id);


--
-- Name: chat_widgets chat_widgets_tenant_id_project_id_domain_id_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_widgets
    ADD CONSTRAINT chat_widgets_tenant_id_project_id_domain_id_key UNIQUE (tenant_id, project_id, domain_id);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: customers customers_tenant_id_email_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_tenant_id_email_key UNIQUE (tenant_id, email);


--
-- Name: email_attachments email_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_pkey PRIMARY KEY (id);


--
-- Name: email_connectors email_connectors_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_connectors
    ADD CONSTRAINT email_connectors_pkey PRIMARY KEY (id);


--
-- Name: email_domain_validations email_domain_validations_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_domain_validations
    ADD CONSTRAINT email_domain_validations_pkey PRIMARY KEY (id);


--
-- Name: email_domain_validations email_domain_validations_tenant_id_project_id_domain_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_domain_validations
    ADD CONSTRAINT email_domain_validations_tenant_id_project_id_domain_key UNIQUE (tenant_id, project_id, domain);


--
-- Name: email_inbox email_inbox_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_pkey PRIMARY KEY (id);


--
-- Name: email_inbox email_inbox_tenant_id_message_id_mailbox_address_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_tenant_id_message_id_mailbox_address_key UNIQUE (tenant_id, message_id, mailbox_address);


--
-- Name: email_mailboxes email_mailboxes_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_mailboxes
    ADD CONSTRAINT email_mailboxes_pkey PRIMARY KEY (id);


--
-- Name: email_mailboxes email_mailboxes_tenant_id_address_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_mailboxes
    ADD CONSTRAINT email_mailboxes_tenant_id_address_key UNIQUE (tenant_id, address);


--
-- Name: email_sync_status email_sync_status_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_sync_status
    ADD CONSTRAINT email_sync_status_pkey PRIMARY KEY (id);


--
-- Name: email_sync_status email_sync_status_tenant_id_connector_id_mailbox_address_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_sync_status
    ADD CONSTRAINT email_sync_status_tenant_id_connector_id_mailbox_address_key UNIQUE (tenant_id, connector_id, mailbox_address);


--
-- Name: email_transports email_transports_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_transports
    ADD CONSTRAINT email_transports_pkey PRIMARY KEY (id);


--
-- Name: file_attachments file_attachments_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.file_attachments
    ADD CONSTRAINT file_attachments_pkey PRIMARY KEY (id);


--
-- Name: goose_db_version goose_db_version_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.goose_db_version
    ADD CONSTRAINT goose_db_version_pkey PRIMARY KEY (id);


--
-- Name: integration_categories integration_categories_name_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_categories
    ADD CONSTRAINT integration_categories_name_key UNIQUE (name);


--
-- Name: integration_categories integration_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_categories
    ADD CONSTRAINT integration_categories_pkey PRIMARY KEY (id);


--
-- Name: integration_sync_logs integration_sync_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_sync_logs
    ADD CONSTRAINT integration_sync_logs_pkey PRIMARY KEY (id);


--
-- Name: integration_templates integration_templates_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_templates
    ADD CONSTRAINT integration_templates_pkey PRIMARY KEY (id);


--
-- Name: integration_templates integration_templates_type_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_templates
    ADD CONSTRAINT integration_templates_type_key UNIQUE (type);


--
-- Name: integrations integrations_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_pkey PRIMARY KEY (id);


--
-- Name: notification_deliveries notification_deliveries_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notification_deliveries
    ADD CONSTRAINT notification_deliveries_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: oauth_tokens oauth_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.oauth_tokens
    ADD CONSTRAINT oauth_tokens_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: projects projects_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_pkey PRIMARY KEY (id);


--
-- Name: projects projects_tenant_id_key_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_tenant_id_key_key UNIQUE (tenant_id, key);


--
-- Name: rate_limit_buckets rate_limit_buckets_identifier_bucket_type_window_start_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.rate_limit_buckets
    ADD CONSTRAINT rate_limit_buckets_identifier_bucket_type_window_start_key UNIQUE (identifier, bucket_type, window_start);


--
-- Name: rate_limit_buckets rate_limit_buckets_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.rate_limit_buckets
    ADD CONSTRAINT rate_limit_buckets_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (role, perm);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (role);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: tenant_project_settings tenant_project_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tenant_project_settings
    ADD CONSTRAINT tenant_project_settings_pkey PRIMARY KEY (id);


--
-- Name: tenant_project_settings tenant_project_settings_tenant_id_project_id_setting_key_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tenant_project_settings
    ADD CONSTRAINT tenant_project_settings_tenant_id_project_id_setting_key_key UNIQUE (tenant_id, project_id, setting_key);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: ticket_mail_routing ticket_mail_routing_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_mail_routing
    ADD CONSTRAINT ticket_mail_routing_pkey PRIMARY KEY (id);


--
-- Name: ticket_mail_routing ticket_mail_routing_tenant_id_ticket_id_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_mail_routing
    ADD CONSTRAINT ticket_mail_routing_tenant_id_ticket_id_key UNIQUE (tenant_id, ticket_id);


--
-- Name: ticket_messages ticket_messages_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_messages
    ADD CONSTRAINT ticket_messages_pkey PRIMARY KEY (id);


--
-- Name: ticket_tags ticket_tags_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_tags
    ADD CONSTRAINT ticket_tags_pkey PRIMARY KEY (ticket_id, tag);


--
-- Name: tickets tickets_pkey; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_pkey PRIMARY KEY (id);


--
-- Name: tickets tickets_tenant_id_number_key; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_tenant_id_number_key UNIQUE (tenant_id, number);


--
-- Name: api_keys unique_api_key_name_per_tenant; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT unique_api_key_name_per_tenant UNIQUE (tenant_id, name);


--
-- Name: integrations unique_integration_per_project; Type: CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT unique_integration_per_project UNIQUE (tenant_id, project_id, type, name);


--
-- Name: idx_agent_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_agent_tenant_id ON public.agent_project_roles USING btree (agent_id, tenant_id);


--
-- Name: idx_agents_email; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_agents_email ON public.agents USING btree (email);


--
-- Name: idx_agents_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_agents_status ON public.agents USING btree (status);


--
-- Name: idx_agents_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_agents_tenant_id ON public.agents USING btree (tenant_id);


--
-- Name: idx_api_keys_active; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_api_keys_active ON public.api_keys USING btree (is_active) WHERE (is_active = true);


--
-- Name: idx_api_keys_key_hash; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_api_keys_key_hash ON public.api_keys USING btree (key_hash);


--
-- Name: idx_api_keys_project_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_api_keys_project_id ON public.api_keys USING btree (project_id) WHERE (project_id IS NOT NULL);


--
-- Name: idx_api_keys_tenant_active; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_api_keys_tenant_active ON public.api_keys USING btree (tenant_id, is_active) WHERE (is_active = true);


--
-- Name: idx_api_keys_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_api_keys_tenant_id ON public.api_keys USING btree (tenant_id);


--
-- Name: idx_attachments_ticket_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_attachments_ticket_id ON public.attachments USING btree (ticket_id);


--
-- Name: idx_chat_messages_agent_types; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_agent_types ON public.chat_messages USING btree (session_id, author_type, created_at);


--
-- Name: idx_chat_messages_ai_agent; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_ai_agent ON public.chat_messages USING btree (session_id, created_at) WHERE ((author_type)::text = 'ai-agent'::text);


--
-- Name: idx_chat_messages_author; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_author ON public.chat_messages USING btree (author_type, author_id);


--
-- Name: idx_chat_messages_session; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_session ON public.chat_messages USING btree (session_id, created_at);


--
-- Name: idx_chat_messages_unread_agent; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_unread_agent ON public.chat_messages USING btree (session_id, read_by_agent) WHERE ((author_type)::text = 'visitor'::text);


--
-- Name: idx_chat_messages_unread_visitor; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_messages_unread_visitor ON public.chat_messages USING btree (session_id, read_by_visitor) WHERE ((author_type)::text = 'agent'::text);


--
-- Name: idx_chat_participants_agent; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_participants_agent ON public.chat_session_participants USING btree (agent_id, left_at);


--
-- Name: idx_chat_sessions_assigned_agent; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_sessions_assigned_agent ON public.chat_sessions USING btree (assigned_agent_id) WHERE (assigned_agent_id IS NOT NULL);


--
-- Name: idx_chat_sessions_customer; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_sessions_customer ON public.chat_sessions USING btree (customer_id) WHERE (customer_id IS NOT NULL);


--
-- Name: idx_chat_sessions_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_sessions_status ON public.chat_sessions USING btree (status, last_activity_at);


--
-- Name: idx_chat_sessions_tenant_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_sessions_tenant_project ON public.chat_sessions USING btree (tenant_id, project_id);


--
-- Name: idx_chat_sessions_widget; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_sessions_widget ON public.chat_sessions USING btree (widget_id);


--
-- Name: idx_chat_widgets_domain; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_widgets_domain ON public.chat_widgets USING btree (domain_id) WHERE (is_active = true);


--
-- Name: idx_chat_widgets_shape_active; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_widgets_shape_active ON public.chat_widgets USING btree (widget_shape, is_active);


--
-- Name: idx_chat_widgets_tenant_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_chat_widgets_tenant_project ON public.chat_widgets USING btree (tenant_id, project_id);


--
-- Name: idx_client_session_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE UNIQUE INDEX idx_client_session_id ON public.chat_sessions USING btree (client_session_id) WITH (deduplicate_items='true') WHERE (client_session_id IS NOT NULL);


--
-- Name: idx_customers_email; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_customers_email ON public.customers USING btree (email);


--
-- Name: idx_customers_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_customers_tenant_id ON public.customers USING btree (tenant_id);


--
-- Name: idx_email_attachments_email; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_attachments_email ON public.email_attachments USING btree (email_id);


--
-- Name: idx_email_connectors_project_validation; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_connectors_project_validation ON public.email_connectors USING btree (project_id, validation_status) WHERE (project_id IS NOT NULL);


--
-- Name: idx_email_connectors_tenant; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_connectors_tenant ON public.email_connectors USING btree (tenant_id) WHERE (is_active = true);


--
-- Name: idx_email_connectors_tenant_project_name; Type: INDEX; Schema: public; Owner: tms
--

CREATE UNIQUE INDEX idx_email_connectors_tenant_project_name ON public.email_connectors USING btree (tenant_id, project_id, name) WHERE (project_id IS NOT NULL);


--
-- Name: idx_email_domain_validations_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_domain_validations_status ON public.email_domain_validations USING btree (status, expires_at);


--
-- Name: idx_email_domain_validations_tenant_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_domain_validations_tenant_project ON public.email_domain_validations USING btree (tenant_id, project_id);


--
-- Name: idx_email_inbox_is_read; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_is_read ON public.email_inbox USING btree (tenant_id, is_read);


--
-- Name: idx_email_inbox_is_reply; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_is_reply ON public.email_inbox USING btree (tenant_id, is_reply);


--
-- Name: idx_email_inbox_mailbox_received; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_mailbox_received ON public.email_inbox USING btree (mailbox_address, received_at DESC);


--
-- Name: idx_email_inbox_sync_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_sync_status ON public.email_inbox USING btree (tenant_id, sync_status);


--
-- Name: idx_email_inbox_tenant_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_tenant_project ON public.email_inbox USING btree (tenant_id, project_id);


--
-- Name: idx_email_inbox_thread; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_thread ON public.email_inbox USING btree (tenant_id, thread_id) WHERE (thread_id IS NOT NULL);


--
-- Name: idx_email_inbox_ticket; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_inbox_ticket ON public.email_inbox USING btree (ticket_id) WHERE (ticket_id IS NOT NULL);


--
-- Name: idx_email_mailboxes_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_mailboxes_project ON public.email_mailboxes USING btree (project_id) WHERE (project_id IS NOT NULL);


--
-- Name: idx_email_mailboxes_tenant; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_mailboxes_tenant ON public.email_mailboxes USING btree (tenant_id);


--
-- Name: idx_email_mailboxes_tenant_project_address; Type: INDEX; Schema: public; Owner: tms
--

CREATE UNIQUE INDEX idx_email_mailboxes_tenant_project_address ON public.email_mailboxes USING btree (tenant_id, project_id, address) WHERE (project_id IS NOT NULL);


--
-- Name: idx_email_sync_status_connector; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_email_sync_status_connector ON public.email_sync_status USING btree (connector_id, mailbox_address);


--
-- Name: idx_file_attachments_entity; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_file_attachments_entity ON public.file_attachments USING btree (related_entity_type, related_entity_id);


--
-- Name: idx_file_attachments_expires_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_file_attachments_expires_at ON public.file_attachments USING btree (expires_at) WHERE (expires_at IS NOT NULL);


--
-- Name: idx_file_attachments_tenant_project; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_file_attachments_tenant_project ON public.file_attachments USING btree (tenant_id, project_id);


--
-- Name: idx_file_attachments_type; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_file_attachments_type ON public.file_attachments USING btree (attachment_type);


--
-- Name: idx_integration_sync_logs_created_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integration_sync_logs_created_at ON public.integration_sync_logs USING btree (created_at);


--
-- Name: idx_integration_sync_logs_integration_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integration_sync_logs_integration_id ON public.integration_sync_logs USING btree (integration_id);


--
-- Name: idx_integration_sync_logs_operation; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integration_sync_logs_operation ON public.integration_sync_logs USING btree (operation);


--
-- Name: idx_integration_sync_logs_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integration_sync_logs_status ON public.integration_sync_logs USING btree (status);


--
-- Name: idx_integrations_last_sync_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integrations_last_sync_at ON public.integrations USING btree (last_sync_at);


--
-- Name: idx_integrations_project_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integrations_project_id ON public.integrations USING btree (project_id);


--
-- Name: idx_integrations_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integrations_status ON public.integrations USING btree (status);


--
-- Name: idx_integrations_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integrations_tenant_id ON public.integrations USING btree (tenant_id);


--
-- Name: idx_integrations_type; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_integrations_type ON public.integrations USING btree (type);


--
-- Name: idx_notification_deliveries_channel_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notification_deliveries_channel_status ON public.notification_deliveries USING btree (channel, status);


--
-- Name: idx_notification_deliveries_notification; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notification_deliveries_notification ON public.notification_deliveries USING btree (notification_id);


--
-- Name: idx_notifications_expires_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notifications_expires_at ON public.notifications USING btree (expires_at) WHERE (expires_at IS NOT NULL);


--
-- Name: idx_notifications_is_read; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notifications_is_read ON public.notifications USING btree (is_read, created_at);


--
-- Name: idx_notifications_recipient; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notifications_recipient ON public.notifications USING btree (agent_id);


--
-- Name: idx_notifications_type_priority; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_notifications_type_priority ON public.notifications USING btree (type, priority);


--
-- Name: idx_projects_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_projects_status ON public.projects USING btree (status);


--
-- Name: idx_projects_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_projects_tenant_id ON public.projects USING btree (tenant_id);


--
-- Name: idx_rate_limit_buckets_identifier; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_rate_limit_buckets_identifier ON public.rate_limit_buckets USING btree (identifier, bucket_type);


--
-- Name: idx_rate_limit_buckets_window; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_rate_limit_buckets_window ON public.rate_limit_buckets USING btree (window_start, window_duration);


--
-- Name: idx_tenants_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tenants_status ON public.tenants USING btree (status);


--
-- Name: idx_ticket_mail_routing_ticket; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_ticket_mail_routing_ticket ON public.ticket_mail_routing USING btree (tenant_id, ticket_id) WHERE (revoked_at IS NULL);


--
-- Name: idx_ticket_mail_routing_token; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_ticket_mail_routing_token ON public.ticket_mail_routing USING btree (public_token) WHERE (revoked_at IS NULL);


--
-- Name: idx_ticket_messages_created_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_ticket_messages_created_at ON public.ticket_messages USING btree (created_at);


--
-- Name: idx_ticket_messages_ticket_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_ticket_messages_ticket_id ON public.ticket_messages USING btree (ticket_id);


--
-- Name: idx_tickets_assignee_agent_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_assignee_agent_id ON public.tickets USING btree (assignee_agent_id);


--
-- Name: idx_tickets_created_at; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_created_at ON public.tickets USING btree (created_at);


--
-- Name: idx_tickets_project_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_project_id ON public.tickets USING btree (project_id);


--
-- Name: idx_tickets_requester_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_requester_id ON public.tickets USING btree (customer_id);


--
-- Name: idx_tickets_status; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_status ON public.tickets USING btree (status);


--
-- Name: idx_tickets_tenant_id; Type: INDEX; Schema: public; Owner: tms
--

CREATE INDEX idx_tickets_tenant_id ON public.tickets USING btree (tenant_id);


--
-- Name: agent_project_roles trigger_agent_project_roles_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_agent_project_roles_updated_at BEFORE UPDATE ON public.agent_project_roles FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: agents trigger_agents_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_agents_updated_at BEFORE UPDATE ON public.agents FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: file_attachments trigger_file_attachments_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_file_attachments_updated_at BEFORE UPDATE ON public.file_attachments FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: integrations trigger_integrations_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_integrations_updated_at BEFORE UPDATE ON public.integrations FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: notification_deliveries trigger_notification_deliveries_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_notification_deliveries_updated_at BEFORE UPDATE ON public.notification_deliveries FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: notifications trigger_notifications_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_notifications_updated_at BEFORE UPDATE ON public.notifications FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: projects trigger_projects_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_projects_updated_at BEFORE UPDATE ON public.projects FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: rate_limit_buckets trigger_rate_limit_buckets_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_rate_limit_buckets_updated_at BEFORE UPDATE ON public.rate_limit_buckets FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: tickets trigger_set_ticket_number; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_set_ticket_number BEFORE INSERT ON public.tickets FOR EACH ROW EXECUTE FUNCTION public.set_ticket_number();


--
-- Name: tenants trigger_tenants_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_tenants_updated_at BEFORE UPDATE ON public.tenants FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: tickets trigger_tickets_updated_at; Type: TRIGGER; Schema: public; Owner: tms
--

CREATE TRIGGER trigger_tickets_updated_at BEFORE UPDATE ON public.tickets FOR EACH ROW EXECUTE FUNCTION public.update_updated_at_column();


--
-- Name: agent_project_roles agent_project_roles_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agent_project_roles
    ADD CONSTRAINT agent_project_roles_agent_id_fkey FOREIGN KEY (agent_id) REFERENCES public.agents(id) ON DELETE CASCADE;


--
-- Name: agent_project_roles agent_project_roles_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agent_project_roles
    ADD CONSTRAINT agent_project_roles_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: agent_project_roles agent_project_roles_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agent_project_roles
    ADD CONSTRAINT agent_project_roles_role_fkey FOREIGN KEY (role) REFERENCES public.roles(role);


--
-- Name: agent_project_roles agent_project_roles_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agent_project_roles
    ADD CONSTRAINT agent_project_roles_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: agents agents_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.agents
    ADD CONSTRAINT agents_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: api_keys api_keys_created_by_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_created_by_fkey FOREIGN KEY (created_by) REFERENCES public.agents(id);


--
-- Name: api_keys api_keys_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: api_keys api_keys_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: attachments attachments_message_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_message_id_fkey FOREIGN KEY (message_id) REFERENCES public.ticket_messages(id) ON DELETE CASCADE;


--
-- Name: attachments attachments_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: attachments attachments_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: attachments attachments_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.attachments
    ADD CONSTRAINT attachments_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE CASCADE;


--
-- Name: chat_messages chat_messages_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: chat_messages chat_messages_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.chat_sessions(id) ON DELETE CASCADE;


--
-- Name: chat_messages chat_messages_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_messages
    ADD CONSTRAINT chat_messages_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: chat_session_participants chat_session_participants_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_session_participants
    ADD CONSTRAINT chat_session_participants_agent_id_fkey FOREIGN KEY (agent_id) REFERENCES public.agents(id) ON DELETE CASCADE;


--
-- Name: chat_session_participants chat_session_participants_session_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_session_participants
    ADD CONSTRAINT chat_session_participants_session_id_fkey FOREIGN KEY (session_id) REFERENCES public.chat_sessions(id) ON DELETE CASCADE;


--
-- Name: chat_session_participants chat_session_participants_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_session_participants
    ADD CONSTRAINT chat_session_participants_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: chat_sessions chat_sessions_assigned_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_assigned_agent_id_fkey FOREIGN KEY (assigned_agent_id) REFERENCES public.agents(id) ON DELETE SET NULL;


--
-- Name: chat_sessions chat_sessions_customer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_customer_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id) ON DELETE SET NULL;


--
-- Name: chat_sessions chat_sessions_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: chat_sessions chat_sessions_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: chat_sessions chat_sessions_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE SET NULL;


--
-- Name: chat_sessions chat_sessions_widget_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_sessions
    ADD CONSTRAINT chat_sessions_widget_id_fkey FOREIGN KEY (widget_id) REFERENCES public.chat_widgets(id) ON DELETE CASCADE;


--
-- Name: chat_widgets chat_widgets_domain_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_widgets
    ADD CONSTRAINT chat_widgets_domain_id_fkey FOREIGN KEY (domain_id) REFERENCES public.email_domain_validations(id) ON DELETE CASCADE;


--
-- Name: chat_widgets chat_widgets_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_widgets
    ADD CONSTRAINT chat_widgets_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: chat_widgets chat_widgets_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.chat_widgets
    ADD CONSTRAINT chat_widgets_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: customers customers_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_attachments email_attachments_email_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_email_id_fkey FOREIGN KEY (email_id) REFERENCES public.email_inbox(id) ON DELETE CASCADE;


--
-- Name: email_attachments email_attachments_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_attachments
    ADD CONSTRAINT email_attachments_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_connectors email_connectors_oauth_token_ref_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_connectors
    ADD CONSTRAINT email_connectors_oauth_token_ref_fkey FOREIGN KEY (oauth_token_ref) REFERENCES public.oauth_tokens(id);


--
-- Name: email_connectors email_connectors_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_connectors
    ADD CONSTRAINT email_connectors_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: email_connectors email_connectors_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_connectors
    ADD CONSTRAINT email_connectors_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_domain_validations email_domain_validations_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_domain_validations
    ADD CONSTRAINT email_domain_validations_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: email_domain_validations email_domain_validations_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_domain_validations
    ADD CONSTRAINT email_domain_validations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_inbox email_inbox_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.email_connectors(id) ON DELETE CASCADE;


--
-- Name: email_inbox email_inbox_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: email_inbox email_inbox_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_inbox email_inbox_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_inbox
    ADD CONSTRAINT email_inbox_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE SET NULL;


--
-- Name: email_mailboxes email_mailboxes_inbound_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_mailboxes
    ADD CONSTRAINT email_mailboxes_inbound_connector_id_fkey FOREIGN KEY (inbound_connector_id) REFERENCES public.email_connectors(id) ON DELETE CASCADE;


--
-- Name: email_mailboxes email_mailboxes_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_mailboxes
    ADD CONSTRAINT email_mailboxes_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: email_mailboxes email_mailboxes_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_mailboxes
    ADD CONSTRAINT email_mailboxes_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_sync_status email_sync_status_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_sync_status
    ADD CONSTRAINT email_sync_status_connector_id_fkey FOREIGN KEY (connector_id) REFERENCES public.email_connectors(id) ON DELETE CASCADE;


--
-- Name: email_sync_status email_sync_status_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_sync_status
    ADD CONSTRAINT email_sync_status_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: email_transports email_transports_outbound_connector_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_transports
    ADD CONSTRAINT email_transports_outbound_connector_id_fkey FOREIGN KEY (outbound_connector_id) REFERENCES public.email_connectors(id) ON DELETE CASCADE;


--
-- Name: email_transports email_transports_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.email_transports
    ADD CONSTRAINT email_transports_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: file_attachments file_attachments_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.file_attachments
    ADD CONSTRAINT file_attachments_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: file_attachments file_attachments_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.file_attachments
    ADD CONSTRAINT file_attachments_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: file_attachments file_attachments_uploaded_by_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.file_attachments
    ADD CONSTRAINT file_attachments_uploaded_by_agent_id_fkey FOREIGN KEY (uploaded_by_agent_id) REFERENCES public.agents(id) ON DELETE SET NULL;


--
-- Name: customers fk_customers_org_id; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT fk_customers_org_id FOREIGN KEY (org_id) REFERENCES public.organizations(id) ON DELETE SET NULL;


--
-- Name: integrations fk_integrations_oauth_token_id; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT fk_integrations_oauth_token_id FOREIGN KEY (oauth_token_id) REFERENCES public.oauth_tokens(id) ON DELETE SET NULL;


--
-- Name: integration_sync_logs integration_sync_logs_integration_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_sync_logs
    ADD CONSTRAINT integration_sync_logs_integration_id_fkey FOREIGN KEY (integration_id) REFERENCES public.integrations(id) ON DELETE CASCADE;


--
-- Name: integration_sync_logs integration_sync_logs_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_sync_logs
    ADD CONSTRAINT integration_sync_logs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: integration_sync_logs integration_sync_logs_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_sync_logs
    ADD CONSTRAINT integration_sync_logs_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: integration_templates integration_templates_category_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integration_templates
    ADD CONSTRAINT integration_templates_category_id_fkey FOREIGN KEY (category_id) REFERENCES public.integration_categories(id) ON DELETE CASCADE;


--
-- Name: integrations integrations_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: integrations integrations_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.integrations
    ADD CONSTRAINT integrations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: notification_deliveries notification_deliveries_notification_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notification_deliveries
    ADD CONSTRAINT notification_deliveries_notification_id_fkey FOREIGN KEY (notification_id) REFERENCES public.notifications(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_recipient_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_recipient_agent_id_fkey FOREIGN KEY (agent_id) REFERENCES public.agents(id) ON DELETE CASCADE;


--
-- Name: notifications notifications_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: oauth_tokens oauth_tokens_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.oauth_tokens
    ADD CONSTRAINT oauth_tokens_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: organizations organizations_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: projects projects_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.projects
    ADD CONSTRAINT projects_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: rate_limit_buckets rate_limit_buckets_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.rate_limit_buckets
    ADD CONSTRAINT rate_limit_buckets_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: role_permissions role_permissions_role_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_role_fkey FOREIGN KEY (role) REFERENCES public.roles(role) ON DELETE CASCADE;


--
-- Name: tenant_project_settings tenant_project_settings_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tenant_project_settings
    ADD CONSTRAINT tenant_project_settings_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: tenant_project_settings tenant_project_settings_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tenant_project_settings
    ADD CONSTRAINT tenant_project_settings_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: ticket_mail_routing ticket_mail_routing_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_mail_routing
    ADD CONSTRAINT ticket_mail_routing_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: ticket_mail_routing ticket_mail_routing_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_mail_routing
    ADD CONSTRAINT ticket_mail_routing_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: ticket_mail_routing ticket_mail_routing_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_mail_routing
    ADD CONSTRAINT ticket_mail_routing_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE CASCADE;


--
-- Name: ticket_messages ticket_messages_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_messages
    ADD CONSTRAINT ticket_messages_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: ticket_messages ticket_messages_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_messages
    ADD CONSTRAINT ticket_messages_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: ticket_messages ticket_messages_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_messages
    ADD CONSTRAINT ticket_messages_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE CASCADE;


--
-- Name: ticket_tags ticket_tags_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_tags
    ADD CONSTRAINT ticket_tags_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: ticket_tags ticket_tags_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_tags
    ADD CONSTRAINT ticket_tags_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: ticket_tags ticket_tags_ticket_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.ticket_tags
    ADD CONSTRAINT ticket_tags_ticket_id_fkey FOREIGN KEY (ticket_id) REFERENCES public.tickets(id) ON DELETE CASCADE;


--
-- Name: tickets tickets_assignee_agent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_assignee_agent_id_fkey FOREIGN KEY (assignee_agent_id) REFERENCES public.agents(id) ON DELETE SET NULL;


--
-- Name: tickets tickets_project_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id) ON DELETE CASCADE;


--
-- Name: tickets tickets_requester_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_requester_id_fkey FOREIGN KEY (customer_id) REFERENCES public.customers(id);


--
-- Name: tickets tickets_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: tms
--

ALTER TABLE ONLY public.tickets
    ADD CONSTRAINT tickets_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: agents; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.agents ENABLE ROW LEVEL SECURITY;

--
-- Name: agents agents_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY agents_tenant_policy ON public.agents USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: api_keys; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.api_keys ENABLE ROW LEVEL SECURITY;

--
-- Name: api_keys api_keys_tenant_isolation; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY api_keys_tenant_isolation ON public.api_keys USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: attachments; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.attachments ENABLE ROW LEVEL SECURITY;

--
-- Name: attachments attachments_project_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY attachments_project_policy ON public.attachments USING ((project_id = ANY ((string_to_array(current_setting('app.project_ids'::text, true), ','::text))::uuid[])));


--
-- Name: attachments attachments_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY attachments_tenant_policy ON public.attachments USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: chat_messages; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.chat_messages ENABLE ROW LEVEL SECURITY;

--
-- Name: chat_messages chat_messages_tenant_isolation; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY chat_messages_tenant_isolation ON public.chat_messages USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: chat_session_participants chat_participants_tenant_isolation; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY chat_participants_tenant_isolation ON public.chat_session_participants USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: chat_session_participants; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.chat_session_participants ENABLE ROW LEVEL SECURITY;

--
-- Name: chat_sessions; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.chat_sessions ENABLE ROW LEVEL SECURITY;

--
-- Name: chat_sessions chat_sessions_tenant_isolation; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY chat_sessions_tenant_isolation ON public.chat_sessions USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: chat_widgets; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.chat_widgets ENABLE ROW LEVEL SECURITY;

--
-- Name: chat_widgets chat_widgets_tenant_isolation; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY chat_widgets_tenant_isolation ON public.chat_widgets USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: customers; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.customers ENABLE ROW LEVEL SECURITY;

--
-- Name: customers customers_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY customers_tenant_policy ON public.customers USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: email_attachments; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_attachments ENABLE ROW LEVEL SECURITY;

--
-- Name: email_attachments email_attachments_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_attachments_tenant ON public.email_attachments USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: email_connectors; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_connectors ENABLE ROW LEVEL SECURITY;

--
-- Name: email_connectors email_connectors_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_connectors_tenant ON public.email_connectors USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: email_inbox; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_inbox ENABLE ROW LEVEL SECURITY;

--
-- Name: email_inbox email_inbox_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_inbox_tenant ON public.email_inbox USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: email_mailboxes; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_mailboxes ENABLE ROW LEVEL SECURITY;

--
-- Name: email_mailboxes email_mailboxes_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_mailboxes_tenant ON public.email_mailboxes USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: email_sync_status; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_sync_status ENABLE ROW LEVEL SECURITY;

--
-- Name: email_sync_status email_sync_status_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_sync_status_tenant ON public.email_sync_status USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: email_transports; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.email_transports ENABLE ROW LEVEL SECURITY;

--
-- Name: email_transports email_transports_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY email_transports_tenant ON public.email_transports USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: file_attachments; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.file_attachments ENABLE ROW LEVEL SECURITY;

--
-- Name: file_attachments file_attachments_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY file_attachments_tenant_policy ON public.file_attachments USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: integration_sync_logs; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.integration_sync_logs ENABLE ROW LEVEL SECURITY;

--
-- Name: integration_sync_logs integration_sync_logs_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY integration_sync_logs_tenant_policy ON public.integration_sync_logs USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: integrations; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.integrations ENABLE ROW LEVEL SECURITY;

--
-- Name: integrations integrations_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY integrations_tenant_policy ON public.integrations USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: notification_deliveries; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.notification_deliveries ENABLE ROW LEVEL SECURITY;

--
-- Name: notification_deliveries notification_deliveries_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY notification_deliveries_tenant_policy ON public.notification_deliveries USING ((EXISTS ( SELECT 1
   FROM public.notifications n
  WHERE ((n.id = notification_deliveries.notification_id) AND (n.tenant_id = (current_setting('app.current_tenant_id'::text))::uuid)))));


--
-- Name: notifications; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.notifications ENABLE ROW LEVEL SECURITY;

--
-- Name: notifications notifications_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY notifications_tenant_policy ON public.notifications USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: oauth_tokens; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.oauth_tokens ENABLE ROW LEVEL SECURITY;

--
-- Name: oauth_tokens oauth_tokens_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY oauth_tokens_tenant ON public.oauth_tokens USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: organizations; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.organizations ENABLE ROW LEVEL SECURITY;

--
-- Name: organizations organizations_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY organizations_tenant_policy ON public.organizations USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: projects; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.projects ENABLE ROW LEVEL SECURITY;

--
-- Name: projects projects_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY projects_tenant_policy ON public.projects USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: rate_limit_buckets; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.rate_limit_buckets ENABLE ROW LEVEL SECURITY;

--
-- Name: rate_limit_buckets rate_limit_buckets_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY rate_limit_buckets_tenant_policy ON public.rate_limit_buckets USING ((tenant_id = (current_setting('app.current_tenant_id'::text))::uuid));


--
-- Name: tenants; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.tenants ENABLE ROW LEVEL SECURITY;

--
-- Name: tenants tenants_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY tenants_tenant_policy ON public.tenants USING ((id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: ticket_mail_routing; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.ticket_mail_routing ENABLE ROW LEVEL SECURITY;

--
-- Name: ticket_mail_routing ticket_mail_routing_tenant; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY ticket_mail_routing_tenant ON public.ticket_mail_routing USING ((tenant_id = (current_setting('app.tenant_id'::text))::uuid));


--
-- Name: ticket_messages; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.ticket_messages ENABLE ROW LEVEL SECURITY;

--
-- Name: ticket_messages ticket_messages_project_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY ticket_messages_project_policy ON public.ticket_messages USING ((project_id = ANY ((string_to_array(current_setting('app.project_ids'::text, true), ','::text))::uuid[])));


--
-- Name: ticket_messages ticket_messages_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY ticket_messages_tenant_policy ON public.ticket_messages USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: ticket_tags; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.ticket_tags ENABLE ROW LEVEL SECURITY;

--
-- Name: ticket_tags ticket_tags_project_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY ticket_tags_project_policy ON public.ticket_tags USING ((project_id = ANY ((string_to_array(current_setting('app.project_ids'::text, true), ','::text))::uuid[])));


--
-- Name: ticket_tags ticket_tags_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY ticket_tags_tenant_policy ON public.ticket_tags USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- Name: tickets; Type: ROW SECURITY; Schema: public; Owner: tms
--

ALTER TABLE public.tickets ENABLE ROW LEVEL SECURITY;

--
-- Name: tickets tickets_project_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY tickets_project_policy ON public.tickets USING ((project_id = ANY ((string_to_array(current_setting('app.project_ids'::text, true), ','::text))::uuid[])));


--
-- Name: tickets tickets_tenant_policy; Type: POLICY; Schema: public; Owner: tms
--

CREATE POLICY tickets_tenant_policy ON public.tickets USING ((tenant_id = (current_setting('app.tenant_id'::text, true))::uuid));


--
-- PostgreSQL database dump complete
--

