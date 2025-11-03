.PHONY: demo-1 demo-2 demo-3 demo-4 demo-5 demo-6 demo-7 demo-8 demo-9 demo-10 demo-10-log log log-follow help

help:
	@echo "Available demos:"
	@echo "  make demo-1  - FlexBox Simple"
	@echo "  make demo-2  - FlexBox Horizontal"
	@echo "  make demo-3  - FlexBox with Table"
	@echo "  make demo-4  - Table Simple String"
	@echo "  make demo-5  - Table Multi-Type"
	@echo "  make demo-6  - FlexBox Nested Borders"
	@echo "  make demo-7  - FlexBox Simple Borders"
	@echo "  make demo-8  - FlexBox Fixed Rows"
	@echo "  make demo-9  - FlexBox Fixed Width Columns"
	@echo "  make demo-10 - FlexBox Mixed Fixed Layout"
	@echo ""
	@echo "Logging commands:"
	@echo "  make demo-10-log - Run demo-10 with size logging"
	@echo "  make log         - View the last 50 lines of the log"
	@echo "  make log-follow  - Follow the log in real-time"

demo-1:
	@go run ./example/flex-box-simple/main.go

demo-2:
	@go run ./example/flex-box-horizonal/main.go

demo-3:
	@cd example/flex-box-with-table && go run main.go

demo-4:
	@cd example/table-simple-string && go run main.go

demo-5:
	@cd example/table-multi-type && go run main.go

demo-6:
	@go run ./example/flex-box-nested-borders/main.go

demo-7:
	@go run ./example/flex-box-simple-borders/main.go

demo-8:
	@go run ./example/flex-box-fixed-rows/main.go

demo-9:
	@go run ./example/flex-box-fixed-width/main.go

demo-10:
	@go run ./example/flex-box-mixed-fixed/main.go

demo-10-log:
	@echo "Starting demo-10 with logging to demo10_size_log.txt..."
	@go run ./example/flex-box-mixed-fixed/main_with_log.go

log:
	@if [ -f demo10_size_log.txt ]; then \
		echo "=== Viewing last 50 lines of demo10_size_log.txt ==="; \
		tail -50 demo10_size_log.txt; \
	else \
		echo "No log file found. Run 'make demo-10-log' first."; \
	fi

log-follow:
	@echo "Following demo10_size_log.txt (Ctrl+C to stop)..."
	@touch demo10_size_log.txt
	@tail -f demo10_size_log.txt
