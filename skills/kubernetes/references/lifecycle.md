# Kubernetes object lifecycle

Creation, reconciliation and deletion are asynchronous. An object with a
deletion timestamp may remain while dependants terminate or a finalizer's
controller completes its work. Owner references and garbage collection determine
which dependent objects are removed with an owner.

- Distinguish an object that was never accepted, one that is pending creation,
  one that is terminating and one that is recreated with a new UID.
- For a stuck Terminating object, inspect finalizers, owner/dependent objects,
  status and recent events. Identify the controller responsible for each
  finalizer before changing it.
- Do not remove a finalizer blindly. Remove one only when the protected cleanup
  is complete or the task provides evidence that the controller is permanently
  unable to perform it and the consequence is understood.
- Respect graceful termination, pre-stop behavior, termination grace periods
  and disruption policies when interpreting delayed shutdown.
- Treat force deletion, orphaning dependants and recreating objects as
  disruptive operations that can lose cleanup or data guarantees.

After lifecycle changes, verify the intended owner/dependent graph and the
observable workload state rather than relying on disappearance from a listing.
