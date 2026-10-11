# Pet care

Pet care is a local journal example for Windows and Android. It has a demo
sign-in gate, pet profiles, dated care notes, and schedule statuses based on
the dates entered by the user. It does not provide account security, sync,
medical advice, medication dosing, or background alerts.

The demo email is `demo@petcare.local` and the password is `pawprints`.
Both are filled in when the app opens, but you still need to submit the form.
These values are bundled with the example and do not protect stored data.

Desktop data lives in `os.UserConfigDir()/ownframe/pet-care/pet-care.db`.
Pass `-db /path/to/file.db` to choose a different database. Android stores
data in the app-private files directory under `ownframe/pet-care/pet-care.db`.
The first empty database receives three sample pets and care records across
upcoming, due-today, overdue, and completed states. Existing data is preserved.
If the configured database cannot open, the app warns that the current session
uses temporary in-memory storage.

The shared Go app has Today, My pets, Pet profile, Care schedule, Care history,
and Settings screens. The first screen is the sign-in gate. Pet and care IDs
remain stable in the local SQLite file. Dates use `YYYY-MM-DD`; schedule labels
compare those dates with the device's local calendar day.

Settings exports version 1 JSON. Windows shows it in a text area for copying
and pasting. Android opens the system document picker to save or choose a JSON
file. Restore validates the whole backup before asking for confirmation; a
rejected backup does not change the database.

Run the Windows build with `go run ./examples/pet-care`. See
`android/README.md` to build or install the Android debug APK. The APK is
copied to `temp/pet-care/pet-care-debug.apk`. Build success
does not verify device input, rotation, keyboard, or Back behavior.
