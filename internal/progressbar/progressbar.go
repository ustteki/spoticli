package progressbar

import (
	"fmt"
	"strings"
	"time"
)

type ProgressBar struct {
	width       int
	position    time.Duration
	duration    time.Duration
	isVisible   bool
	isDragging  bool
	dragPos     float64
}

func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		width:     width,
		isVisible: true,
	}
}

func (pb *ProgressBar) Update(position, duration time.Duration) {
	if !pb.isDragging {
		pb.position = position
	}
	pb.duration = duration
}

func (pb *ProgressBar) SetVisible(visible bool) {
	pb.isVisible = visible
}

func (pb *ProgressBar) IsVisible() bool {
	return pb.isVisible
}

func (pb *ProgressBar) SetWidth(width int) {
	pb.width = width
}

func (pb *ProgressBar) Render() string {
	if !pb.isVisible {
		return ""
	}
	
	var result strings.Builder
	
	var progress float64
	if pb.isDragging {
		progress = pb.dragPos
	} else if pb.duration > 0 {
		progress = float64(pb.position) / float64(pb.duration)
	}
	
	if progress > 1.0 {
		progress = 1.0
	} else if progress < 0.0 {
		progress = 0.0
	}
	
	currentTime := pb.position
	totalTime := pb.duration
	
	if pb.isDragging && pb.duration > 0 {
		currentTime = time.Duration(float64(pb.duration) * pb.dragPos)
	}
	
	timeStr := fmt.Sprintf("[dim]%s[white] / [dim]%s[white]", 
		formatDuration(currentTime), 
		formatDuration(totalTime))
	
	timeDisplayLen := len(strings.ReplaceAll(strings.ReplaceAll(timeStr, "[dim]", ""), "[white]", ""))
	barWidth := pb.width - timeDisplayLen - 3 // 3 for spacing
	
	if barWidth < 10 {
		barWidth = 10
	}
	
	filledWidth := int(float64(barWidth) * progress)
	
	result.WriteString(timeStr)
	result.WriteString(" [")
	
	for i := 0; i < barWidth; i++ {
		if i < filledWidth {
			if pb.isDragging {
				result.WriteString("[yellow]█[white]")
			} else {
				result.WriteString("[green]█[white]")
			}
		} else if i == filledWidth && filledWidth < barWidth {
			if pb.isDragging {
				result.WriteString("[yellow]▌[white]")
			} else {
				result.WriteString("[green]▌[white]")
			}
		} else {
			result.WriteString("[dim]░[white]")
		}
	}
	
	result.WriteString("]")
	
	if pb.duration > 0 {
		percentage := int(progress * 100)
		result.WriteString(fmt.Sprintf(" [dim]%d%%[white]", percentage))
	}
	
	return result.String()
}

func (pb *ProgressBar) HandleClick(x, y, maxWidth int) (float64, bool) {
	if !pb.isVisible || pb.duration == 0 {
		return 0, false
	}
	
	timeStr := fmt.Sprintf("%s / %s", 
		formatDuration(pb.position), 
		formatDuration(pb.duration))
	
	timeDisplayLen := len(timeStr)
	barStart := timeDisplayLen + 2 // " [" before the bar
	
	barWidth := maxWidth - timeDisplayLen - 3
	if barWidth < 10 {
		barWidth = 10
	}
	
	if x >= barStart && x < barStart+barWidth {
		clickPos := x - barStart
		percentage := float64(clickPos) / float64(barWidth)
		
		if percentage > 1.0 {
			percentage = 1.0
		} else if percentage < 0.0 {
			percentage = 0.0
		}
		
		return percentage, true
	}
	
	return 0, false
}

func (pb *ProgressBar) StartDrag(position float64) {
	pb.isDragging = true
	pb.dragPos = position
}

func (pb *ProgressBar) UpdateDrag(position float64) {
	if pb.isDragging {
		pb.dragPos = position
		if pb.dragPos > 1.0 {
			pb.dragPos = 1.0
		} else if pb.dragPos < 0.0 {
			pb.dragPos = 0.0
		}
	}
}

func (pb *ProgressBar) EndDrag() (time.Duration, bool) {
	if !pb.isDragging {
		return 0, false
	}
	
	pb.isDragging = false
	if pb.duration > 0 {
		newPosition := time.Duration(float64(pb.duration) * pb.dragPos)
		return newPosition, true
	}
	
	return 0, false
}

func (pb *ProgressBar) IsDragging() bool {
	return pb.isDragging
}

func (pb *ProgressBar) GetProgress() float64 {
	if pb.duration == 0 {
		return 0
	}
	
	if pb.isDragging {
		return pb.dragPos
	}
	
	return float64(pb.position) / float64(pb.duration)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}