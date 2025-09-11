job "ai-agent" {
  datacenters = ["dc1"]
  type        = "service"

  group "ai-agent-service" {
    count = 5  # High availability with 5 replicas

    # Spread across different nodes for better distribution
    # spread {
    #   attribute = "${meta.region}"
    #   target "falkenstein" {
    #     percent = 60
    #   }
    #   target "iowa" {
    #     percent = 40
    #   } 
    # }
    
    constraint {
      attribute = "${attr.kernel.name}"
      value     = "linux"
    }

    constraint {
      attribute = "${meta.region}"
      value     = "iowa"
    }
    
    network {
      mode = "bridge"
      port "http" {
      }
    }

    volume "tms_backend_storage" {
      type      = "host"
      read_only = false
      source    = "tms_backend_storage"
    }
    
    service {
      name = "tms-backend"
      port = "http"
      
      # Health checks
      check {
        type     = "http"
        path     = "/api/health"
        interval = "30s"
        timeout  = "10s"
        
        check_restart {
          limit = 3
          grace = "30s"
          ignore_warnings = false
        }
      }
    }
    
    # Restart policy for resilience
    restart {
      attempts = 5
      interval = "2m"
      delay    = "3s"
      mode     = "fail"
    }
    
    # Rolling update configuration
    # update {
    #   max_parallel      = 1
    #   min_healthy_time  = "30s"
    #   healthy_deadline  = "3m"
    #   progress_deadline = "10m"
    #   auto_revert       = true
    #   auto_promote      = true
    #   canary            = 1
    #   stagger           = "5s"
    # }
    
    # Placement preferences for load distribution
    # affinity {
    #   attribute = "${node.unique.id}"
    #   operator  = "regexp"
    #   value     = ".*"
    #   weight    = 50
    # }
    
    task "backend" {
      driver = "docker"

      volume_mount {
        volume      = "tms_backend_storage"
        destination = "/opt/tms"
        read_only   = false
      }
      
      # Enable Vault workload identity
      # identity {
      #   aud = ["vault.io"]
      #   env = true
      #   file = true
      #   change_mode = "restart"
      # }

      vault {
        policies = ["nomad-cluster"]
      }
      
      template {
        data = <<EOH
REGION_TAG=region={{ env "meta.region" }}
EOH
        destination = "secrets/region.env"
        env         = true
        change_mode = "restart"
      }
      
      # API keys and secrets from Vault
      template {
        data = <<EOH
{{- with secret "secret/data/shared/githubAuth" -}}
GHC_TOKEN={{ .Data.data.GHC_TOKEN }}
GITHUB_USERNAME={{ .Data.data.GITHUB_USERNAME }}
{{- end }}
EOH
        destination = "secrets/github.env"
        env         = true
        change_mode = "restart"
      }
      
      # Application configuration from Vault
      template {
        data = <<EOH
{{- with secret "secret/data/tms/config" -}}
APP_ENV={{ .Data.data.APP_ENV }}
APP_NAME={{ .Data.data.APP_NAME }}
LOG_LEVEL={{ .Data.data.LOG_LEVEL }}
AI_API_KEY={{ .Data.data.AI_API_KEY }}
OPENAI_API_KEY={{ .Data.data.AI_API_KEY }}
AI_AGENT_LOGIN_ACCESS_KEY={{ .Data.data.TMS_API_S2S_KEY }}
TMS_API_BASE_URL=http://{{- range $i, $service := service "backend" -}}{{- if eq $i 0 }}{{ .Address }}:{{ .Port }}{{- end }}{{- end }}
{{- end }}
EOH
        destination = "secrets/config.env"
        env         = true
        change_mode = "restart"
      }
      
      # Consul service discovery configuration
      template {
        data = <<EOH
CONSUL_HTTP_ADDR=http://{{ env "NOMAD_IP_http" }}:8500
SERVICE_NAME=backend
SERVICE_ID=backend-{{ env "NOMAD_ALLOC_ID" }}
SERVER_PORT={{ env "NOMAD_PORT_http" }}
PORT={{ env "NOMAD_PORT_http" }}
EOH
        destination = "secrets/consul.env"
        env         = true
        change_mode = "restart"
      }
      
      config {
        image = "ghcr.io/taral-co/tms/tms-ai-agent:latest"
        ports = ["http"]
        
        # Docker authentication for private registry
        auth {
          username = "${GITHUB_USERNAME}" 
          password = "${GHC_TOKEN}"
          server_address = "ghcr.io"
        }
        
        # Force pull latest image
        force_pull = true
        
        # Wait for database to be ready
        command = "/bin/sh"
        args = [
          "-c",
          <<EOF
echo 'Waiting for database...' &&
echo 'Database is ready' &&
echo 'Starting Backend service...' &&
# sleep 1000 &&  # Ensure PostgreSQL replica is fully initialized
EOF
        ]
      }
      
      # Performance optimizations
      env {
        GOMAXPROCS = "2"
        GOGC = "100"
        GOMEMLIMIT = "450MiB"
        REGION = "${meta.region}"  # or "iowa"
        CORS_ALLOW_CREDENTIALS="false"
        CORS_ORIGINS=""
      }
      
      # Resource allocation matching Docker Swarm config
      resources {
        cpu    = 300   # 0.15 CPU
        memory = 800   # 512MB
      }
      
      # Logs configurationp
      logs {
        max_files     = 10
        max_file_size = 15
      }
    }
  }
}