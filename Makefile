.PHONY: demo-1 demo-2 demo-3 demo-4 demo-5 demo-6 demo-7 demo-8 demo-9 demo-10 demo-10-log demo-11 demo-12 log log-follow help clean

help:
	@echo "Available demos (press 'q' or Ctrl+C to exit):"
	@echo ""
	@echo "Original stickers demos:"
	@echo "  make demo-1  - FlexBox Simple"
	@echo "                 Basic grid of colored cells using ratio-based sizing"
	@echo ""
	@echo "  make demo-2  - FlexBox Horizontal"
	@echo "                 Column-based layout (vertical stacking within columns)"
	@echo ""
	@echo "  make demo-3  - FlexBox with Table"
	@echo "                 Embedded table inside a flexbox layout with CSV data"
	@echo ""
	@echo "  make demo-4  - Table Simple String"
	@echo "                 Navigable table with sorting (arrows, ctrl+s, enter/space)"
	@echo ""
	@echo "  make demo-5  - Table Multi-Type"
	@echo "                 Table with typed columns (arrows, ctrl+s, enter/space)"
	@echo ""
	@echo "New improvements:"
	@echo "  make demo-6  - FlexBox Nested Borders"
	@echo "                 Recursively nested flexboxes with border styles (t)"
	@echo ""
	@echo "  make demo-7  - FlexBox Simple Borders"
	@echo "                 Grid with different border styles (t)"
	@echo ""
	@echo "  make demo-8  - FlexBox Fixed Rows"
	@echo "                 Fixed vs dynamic row heights (t)"
	@echo ""
	@echo "  make demo-9  - FlexBox Fixed Width Columns"
	@echo "                 Fixed vs dynamic column widths (t)"
	@echo ""
	@echo "  make demo-10 - FlexBox Mixed Fixed Layout"
	@echo "                 Combines fixed widths/heights with ratio-based cells (t, f, h)"
	@echo ""
	@echo "  make demo-11 - FlexBox Interactive Cell Config"
	@echo "                 Configure individual cells interactively (A-T, w, h, T, H, ESC)"
	@echo ""
	@echo "  make demo-12 - FlexBox Row Alignment"
	@echo "                 Demonstrates SetRowAlign for varying row widths (a)"
	@echo ""
	@echo "Logging commands:"
	@echo "  make demo-10-log - Run demo-10 with size logging"
	@echo "  make log         - View the last 50 lines of the log"
	@echo "  make log-follow  - Follow the log in real-time"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean       - Remove log files and build artifacts"

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

demo-11:
	@go run ./example/flex-box-cell-config/main.go

demo-12:
	@go run ./example/flex-box-row-align/main.go

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

clean:
	@echo "Cleaning up..."
	@-pkill -f "go-build.*main" 2>/dev/null || true
	@rm -f demo10_size_log.txt
	@rm -f main
	@rm -f flex-box-simple flex-box-horizonal flex-box-nested-borders flex-box-simple-borders
	@rm -f flex-box-fixed-rows flex-box-fixed-width flex-box-mixed-fixed flex-box-cell-config flex-box-row-align
	@rm -f example/flex-box-with-table/flex-box-with-table
	@rm -f example/table-simple-string/table-simple-string
	@rm -f example/table-multi-type/table-multi-type
	@go clean
	@echo "Done."
