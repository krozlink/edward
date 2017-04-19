package tracker

import "sync"

type Task interface {
	Name() string

	State() TaskState
	SetState(TaskState, ...string)

	Updates() <-chan struct{}
	Close()

	Messages() []string

	Child(name string) Task
	Children() []Task
}

type TaskState int

const (
	TaskStateInProgress TaskState = iota
	TaskStateSuccess
	TaskStateWarning
	TaskStateFailed
)

type task struct {
	name     string
	messages []string
	state    TaskState

	childNames []string
	children   map[string]Task

	updates chan struct{}
	mtx     sync.Mutex
}

func NewTask(name string) Task {
	return &task{
		name:     name,
		children: make(map[string]Task),
		updates:  make(chan struct{}, 2),
	}
}

func (t *task) Name() string {
	return t.name
}

func (t *task) State() TaskState {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if len(t.children) == 0 {
		return t.state
	}

	var state = TaskStateSuccess
	for _, n := range t.childNames {
		child := t.children[n]
		switch child.State() {
		case TaskStateFailed:
			return TaskStateFailed
		case TaskStateInProgress:
			state = TaskStateInProgress
		case TaskStateWarning:
			if state != TaskStateInProgress {
				state = TaskStateWarning
			}
		}
	}
	return state
}

func (t *task) Updates() <-chan struct{} {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	return t.updates
}

func (t *task) Close() {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	close(t.updates)
}

func (t *task) SetState(state TaskState, messages ...string) {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	t.state = state
	t.messages = messages

	t.updates <- struct{}{}
}

func (t *task) Messages() []string {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	return t.messages
}

func (t *task) Child(name string) Task {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	if c, ok := t.children[name]; ok {
		return c
	}

	t.childNames = append(t.childNames, name)
	t.children[name] = &task{
		name:     name,
		children: make(map[string]Task),
		updates:  t.updates,
	}

	t.updates <- struct{}{}
	return t.children[name]
}

func (t *task) Children() []Task {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	var children []Task
	for _, c := range t.childNames {
		children = append(children, t.children[c])
	}
	return children
}