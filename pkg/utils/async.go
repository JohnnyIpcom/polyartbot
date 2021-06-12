package utils

import (
	"context"
	"sync"

	"gitlab.com/jonas.jasas/condchan"
)

type AsyncBool struct {
	cond *condchan.CondChan
	flag bool
}

func NewAsyncBool(b bool) *AsyncBool {
	return &AsyncBool{
		cond: condchan.New(&sync.Mutex{}),
		flag: b,
	}
}

func (n *AsyncBool) Await(ctx context.Context, flag bool) <-chan error {
	errCh := make(chan error, 1)

	go func() {
		defer close(errCh)

		n.cond.L.Lock()
		if n.flag != flag {
			n.cond.Select(func(c <-chan struct{}) {
				select {
				case <-c:
					if n.flag == flag {
						errCh <- nil
						return
					}

				case <-ctx.Done():
					errCh <- ctx.Err()
					return
				}
			})
		} else {
			errCh <- nil
		}
		n.cond.L.Unlock()
	}()

	return errCh
}

func (n *AsyncBool) Notify(flag bool) {
	n.cond.L.Lock()
	defer n.cond.L.Unlock()

	n.flag = flag
	n.cond.Broadcast()
}

func (n *AsyncBool) Get() bool {
	n.cond.L.Lock()
	defer n.cond.L.Unlock()

	return n.flag
}
