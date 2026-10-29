package main

type Task struct {
	ID       string
	Title    string
	Category string
	Done     bool
}

type Tasks struct {
	Items  []Task
	Filter string
	Light  bool
}

func (tasks *Tasks) Completed() int {
	var count int
	for _, task := range tasks.Items {
		if task.Done {
			count++
		}
	}
	return count
}

func (tasks *Tasks) Visible() []Task {
	items := []Task{}
	for _, task := range tasks.Items {
		if tasks.Filter == "open" && task.Done {
			continue
		}
		if tasks.Filter == "done" && !task.Done {
			continue
		}
		items = append(items, task)
	}
	return items
}
