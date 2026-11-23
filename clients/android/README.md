# BrainBlitz for Android

BrainBlitz for Android is a native Kotlin and Jetpack Compose app for multiplayer trivia. Create an account or sign in, view your profile and statistics, and choose a category, timer, and player count. Matches include live answer feedback, round recaps, final standings, and replay. Android and Telegram players can join the same match on the same backend using distinct accounts and matching settings. [Telegram client](../telegram-bot/README.md)

The app saves the access token in app-private storage encrypted with a key held by Android Keystore. When the token expires, sign in again. [Android Keystore](https://developer.android.com/privacy-and-security/keystore)

## Run locally

1. Open `clients/android` in Android Studio. Use its bundled JDK and install Android SDK Platform 35 and Build Tools 35.0.0 through SDK Manager if needed. The app supports Android 8.0 (API 26) and newer.
2. Follow the first-start steps to generate a local `.env`, start Docker Compose, and load the demo questions. The default debug build reaches the host gateway from the emulator at `http://10.0.2.2:8088`. Use host port 8088 for this build. [Backend setup](../../docs/run-on-another-computer.md#first-start)
3. Run the `app` configuration on two emulators and sign in with a distinct account on each. Open **Play**, choose the same category and timer, set the total player count to two, then select **Find match**. Both debug builds should show the same match ID. Select **Ready** on both devices.
4. Answer a question on the first device and check its personal feedback while the second player is still choosing. Have both players answer one round, then let one player miss the next deadline. Check the recap for the correct answer and the missed submission. Rotate a device during a round to check that its deadline and submission are retained.
5. Finish the match, compare both sets of standings, and choose **Play again** to return to the previous match settings. For a mixed-client match, repeat with a Telegram player using a separate BrainBlitz account and the same settings.

## Build and test

From `clients/android`, with JDK 17 or newer and the Android SDK configured, run:

```sh
./gradlew :app:assembleDebug :app:testDebugUnitTest :app:lintDebug
```

The debug APK is at `app/build/outputs/apk/debug/app-debug.apk`. Android Studio can install it, or run `adb -s DEVICE_SERIAL install -r app/build/outputs/apk/debug/app-debug.apk` after finding the device serial with `adb devices`. GitHub Actions runs the debug build, unit tests, and lint for Android changes. [Android CI](../../.github/workflows/android.yml)

## Test against staging

Set `brainBlitzStagingUrl` to the staging gateway's HTTPS origin in Gradle properties, or pass it when building the debug APK. Replace the example with your deployment's origin.

```sh
./gradlew :app:assembleDebug -PbrainBlitzStagingUrl=https://staging.example.com
```

Use this APK on physical devices. The configured gateway needs the user signup, login and profile routes, the game options and WebSocket routes, and the question timer route. Review the ingress before testing staging. [Gateway contracts](../../docs/client/android-roadmap.md#existing-backend-contracts)

When reporting a failure, include the APK version or commit, Android version, environment, and match settings. Redact credentials and tokens from logs and screenshots. [Bug report template](../../.github/ISSUE_TEMPLATE/bug_report.yml)

## Matchmaking behavior

**Cancel search** waits for the server to confirm you left the queue. If matchmaking has already created a match, the app opens it. Everyone has one minute to select **Ready** by default. After the deadline, expired matches let players search again. [Android roadmap](../../docs/client/android-roadmap.md)

## Test match recovery

- Select **Ready** on the first emulator, close the app process, and relaunch. Check that it returns to the same match and is still waiting for the other player.
- Answer a question on one emulator while the other player waits. Disconnect and reconnect the first emulator, then check that the deadline and accepted answer are preserved.
- Let the other client finish the match while the first app is closed. Relaunch the first app to recover the standings. Returning to the profile or choosing **Play again** dismisses those results on that device.

The app starts a bounded reconnect when the active network changes or a game socket drops, and resumes reconnecting when it returns to the foreground. Match state comes from the authenticated game service. The app asks you to sign in again when the access token expires, then checks for your match.

## Validation status

The local debug build passed on two Android 10 emulators against Docker Compose. Tests covered two-client and mixed Android–Telegram gameplay, a 15-player draw, session expiry, queue cancellation, ready timeout, missed answers, match recovery, and duplicate-answer prevention. The emulator pass also checked dark mode, rotation, 200% text, countdown labels, and animations. The debug build and 11 unit tests passed, and lint reported 0 errors and 18 warnings. Staging, physical-device and TalkBack validation, and the signed beta build remain open. [Android beta roadmap](../../docs/client/android-roadmap.md#5-prepare-a-beta-build)
