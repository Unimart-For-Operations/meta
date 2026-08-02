module github.com/${{ values.destination.owner }}/${{ values.name }}

go 1.23

require (
	github.com/gofiber/fiber/v2 v2.52.5
	{% if values.enableMetrics %}github.com/gofiber/adaptor/v2 v2.2.1
	github.com/prometheus/client_golang v1.19.1{% endif %}
)
