# Offline mode for the field notes app

## Destination

A field worker can open the app with no signal, read and edit the notes they need for the day, and have every edit reach the server once they are back online, with no silent data loss.

## Notes

An example map that ships with wayfinder-maps so you can try the viewer without a map of your own. The tickets are made up.

## Decisions so far

- [Which sync engines exist for SQLite on phones?](./tickets/01-which-sync-engines-exist-for-sqlite-on-phones.md): three hold up; we take the one with a hosted option and an open protocol.
- [Which data must work offline?](./tickets/02-which-data-must-work-offline.md): notes, checklists and the day's job list; account settings stay online only.

## Not yet specified

- **Merge rules for checklists.** Whether ticking a box is its own change or part of the note. <clears-with: 04>
- **Offline indicator.** How the app says it is offline without nagging.

## Out of scope

- [Does offline cover photo attachments?](./tickets/08-does-offline-cover-photo-attachments.md): photos stay online only for the first release.
