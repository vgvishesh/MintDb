# MintDb

## Code comments

Default to writing no comments. Add one only when it meets one of these rules:

1. **Warn about consequences.** Comment when a reader must know about a non-obvious consequence of the logic before changing or relying on it, such as data loss, ordering requirements, concurrency hazards, or on-disk format assumptions.
   ```go
   // Replay treats any record containing "Delete:" as a tombstone, so a value
   // with that text deletes a key on restart.
   ```
2. **Document abstractions.** Comment the fields of structs and the methods of interfaces that other code depends on, when their meaning or contract isn't clear from the name and type.
3. **Explain tricky code.** Comment a function or block only when it is genuinely hard to follow and a short explanation saves the reader real effort.

Don't write comments that:
- restate what the code already says (`// increment counter`)
- narrate changes or history (`// added X`, `// fixed bug`); that belongs in commit messages
- explain obvious functions, fields, or control flow

When in doubt, leave the comment out, or make the code clearer instead with better names or smaller functions.
