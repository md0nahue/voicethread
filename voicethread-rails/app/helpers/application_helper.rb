module ApplicationHelper
  def format_duration(seconds)
    return "0:00" if seconds.nil? || seconds.zero?

    minutes = (seconds / 60).floor
    remaining_seconds = (seconds % 60).floor
    hours = (minutes / 60).floor
    minutes = minutes % 60

    if hours > 0
      format("%d:%02d:%02d", hours, minutes, remaining_seconds)
    else
      format("%d:%02d", minutes, remaining_seconds)
    end
  end
end
