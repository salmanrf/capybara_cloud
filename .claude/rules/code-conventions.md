# Code conventions

These apply to every sub-project (`apps/*`, `packages/*`), in every language.

## 1. Docstrings over scattered comments

Explain code with a docstring on the function (or type), not with single-line comments sprinkled through the body. Every docstring uses this exact format, in this order:

1. `@params`: what the function accepts.
2. `@return`: what it is expected to return, including errors.
3. The description of what it does.

Go:

```go
// @params user_id: ID of the requesting user; app_id: ID of the application to deploy; bundle: the uploaded .tar.gz
// @return the created deployment, or an error if the app is not found or the user lacks access
// Deploy validates access to the application, stores the bundle and queues the deployment pipeline.
func (s *service) Deploy(user_id string, app_id string, bundle multipart.File) (*database.ApplicationDeployment, error) {
```

TypeScript:

```ts
/**
 * @params id: the application ID
 * @return the query result for the application
 * Fetches a single application and keeps it cached under its query key.
 */
export function useApplication(id: string) {
```

## 2. Readability over cleverness

### 2.1 No one-liners or shortcuts

Always write the block form of control flow, with the body on its own lines. Never write the one-line version.

```ts
// ❌
if (!user) return null;
for (const a of apps) render(a);
while (queue.length) process(queue.shift());

// ✅
if (!user) {
  return null;
}
```

This also applies to anything that squeezes control flow onto one line, such as nested ternaries or `&&`/`||` used as a statement (`ok && doThing()`).

### 2.2 A condition must never span multiple lines

`&&` and `||` are allowed. But if a condition is long enough to wrap onto several lines, it is testing too much, and the condition design is already wrong. Break it into separate `if`s: sequential guard clauses or nested `if`s.

```go
// ❌
if user != nil &&
	user.OrgID == org.ID &&
	slices.Contains(roles, member.Role) {
	...
}

// ✅
if user == nil {
	return ErrUnauthorized
}
if user.OrgID != org.ID {
	return ErrForbidden
}
if !slices.Contains(roles, member.Role) {
	return ErrForbidden
}
...
```
