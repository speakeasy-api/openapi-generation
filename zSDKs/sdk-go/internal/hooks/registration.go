package hooks

func initHooks(h *Hooks) {
	ih := &IdempotencyHook{}

	h.registerBeforeRequestHook(ih)
}
