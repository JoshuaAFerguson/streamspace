# Session Scheduling User Guide

**Difficulty**: Beginner
**Time Required**: 10-15 minutes

Automate your workflow with scheduled sessions that start automatically.

---

## What is Session Scheduling?

Session Scheduling allows you to:
- **Automate**: Sessions start automatically at specified times
- **Repeat**: Daily, weekly, or monthly recurring sessions
- **Integrate**: Sync with Google Calendar or Outlook
- **Optimize**: Pre-warm sessions for instant access
- **Control**: Auto-terminate sessions after a duration

---

## Creating Your First Scheduled Session

### Step 1: Navigate to Scheduling

1. Login to StreamSpace
2. Click **Scheduling** in the left menu
3. Click **New Schedule** button

### Step 2: Configure Basic Settings

**Name Your Schedule**:
- Enter a descriptive name (e.g., "Morning Dev Environment")
- This appears in your schedule list and calendar

**Choose a Template**:
- Select from dropdown (Firefox, VS Code, etc.)
- This determines what application runs

### Step 3: Choose Schedule Type

#### One-time Schedule
- **Use Case**: Special event or one-off task
- **Settings**: Pick date and time
- **Example**: "Client demo on Friday at 2pm"

#### Daily Schedule
- **Use Case**: Routine daily tasks
- **Settings**: Pick time of day
- **Example**: "Start dev environment at 9am every day"

#### Weekly Schedule
- **Use Case**: Specific days each week
- **Settings**:
  - Select days (Mon, Tue, Wed, etc.)
  - Pick time of day
- **Example**: "Team meetings on Mon/Wed/Fri at 10am"

#### Monthly Schedule
- **Use Case**: Monthly reports or reviews
- **Settings**:
  - Pick day of month (1-31)
  - Pick time of day
- **Example**: "First of month at 9am"

#### Cron Expression (Advanced)
- **Use Case**: Complex schedules
- **Settings**: Enter cron expression
- **Example**: `0 */4 * * *` (every 4 hours)
- **Need help?**: Use [crontab.guru](https://crontab.guru)

### Step 4: Set Timezone

- Select your timezone from dropdown
- Sessions will start at the specified time in that timezone
- Example: "America/New_York" for EST/EDT

### Step 5: Configure Auto-termination (Optional)

**Why Use This?**
- Saves resources when you forget to close sessions
- Ensures sessions don't run longer than needed

**Settings**:
- Toggle **Auto-terminate after duration**
- Enter duration in minutes (default: 480 = 8 hours)
- Example: Auto-stop after 2 hours of inactivity

### Step 6: Enable Pre-warming (Optional)

**Why Use This?**
- Session is ready instantly when you need it
- No waiting for container startup

**Settings**:
- Toggle **Pre-warm session before scheduled time**
- Enter minutes before start time (default: 5)
- Example: Start warming up 5 minutes early

### Step 7: Create Schedule

1. Review your settings
2. Click **Create**
3. Your schedule appears in the list

---

## Managing Scheduled Sessions

### View Your Schedules

Navigate to **Scheduling** → **Scheduled Sessions** tab:

**Schedule List Shows**:
- Name
- Schedule pattern (Daily at 9:00, etc.)
- Next run time
- Last run status
- Enable/Disable toggle

### Enable/Disable a Schedule

**Temporarily Stop**:
- Click the Pause icon
- Schedule is disabled but not deleted
- Click Play icon to re-enable

**Permanently Delete**:
- Click the Delete icon (trash can)
- Confirm deletion
- Schedule is removed

### Edit a Schedule

Currently, you cannot edit schedules. To change:
1. Delete the existing schedule
2. Create a new one with updated settings

---

## Calendar Integration

Sync your scheduled sessions to your calendar app.

### Connect Google Calendar

#### Step 1: Navigate to Calendar Integration

1. Go to **Scheduling** → **Calendar Integration** tab
2. Click **Connect Calendar**
3. Select **Google Calendar**

#### Step 2: Authorize Access

1. You'll be redirected to Google
2. Select your Google account
3. Review permissions:
   - Read/write calendar events
   - Access calendar list
4. Click **Allow**

#### Step 3: Verify Connection

1. You'll return to StreamSpace
2. Your Google Calendar appears as connected
3. Sync status shows "Last synced" time

### Connect Outlook Calendar

#### Step 1: Navigate to Calendar Integration

1. Go to **Scheduling** → **Calendar Integration** tab
2. Click **Connect Calendar**
3. Select **Outlook Calendar**

#### Step 2: Authorize Access

1. You'll be redirected to Microsoft
2. Enter your Microsoft account credentials
3. Review permissions and accept

#### Step 3: Verify Connection

1. Return to StreamSpace
2. Outlook Calendar shows as connected

### How Calendar Sync Works

**What Gets Synced**:
- All enabled scheduled sessions
- Session name becomes event title
- Scheduled time becomes event start
- Auto-terminate duration becomes event length

**What Appears in Your Calendar**:
```
Event: Morning Dev Environment
Time: 9:00 AM - 5:00 PM
Location: https://streamspace.local/sessions/user1-vscode
Description: StreamSpace scheduled session (VS Code)
```

**Sync Frequency**:
- Every 15 minutes automatically
- Click **Sync Now** for immediate sync

### Disconnect a Calendar

1. Go to **Scheduling** → **Calendar Integration** tab
2. Find the connected calendar
3. Click **Disconnect**
4. Confirm disconnection
5. Calendar events are removed

---

## Export to iCal

### Why Use iCal Export?

- Works with any calendar app (Apple Calendar, Thunderbird, etc.)
- No OAuth connection required
- One-time export for backup

### Export Steps

1. Go to **Scheduling**
2. Click **Export iCal** button (top right)
3. Save the `.ics` file
4. Import into your calendar app:
   - **Apple Calendar**: File → Import
   - **Outlook**: File → Open & Export → Import
   - **Thunderbird**: Right-click calendar → Import

### When to Re-export

- After creating new schedules
- After deleting schedules
- After changing schedule times

---

## Use Cases & Examples

### Daily Development Environment

**Scenario**: Start coding workspace every weekday

**Schedule**:
- Type: Weekly
- Days: Mon, Tue, Wed, Thu, Fri
- Time: 9:00 AM
- Template: VS Code
- Timezone: America/New_York
- Auto-terminate: 8 hours
- Pre-warm: 5 minutes

### Weekly Team Meeting Workspace

**Scenario**: Firefox browser for video calls every Monday

**Schedule**:
- Type: Weekly
- Days: Monday
- Time: 10:00 AM
- Template: Firefox
- Timezone: UTC
- Auto-terminate: 2 hours
- Pre-warm: 10 minutes (test audio/video early)

### Monthly Reporting Session

**Scenario**: LibreOffice for monthly reports on the 1st

**Schedule**:
- Type: Monthly
- Day of Month: 1
- Time: 8:00 AM
- Template: LibreOffice
- Timezone: Europe/London
- Auto-terminate: 4 hours

### On-demand Client Demo

**Scenario**: One-time demo session for client meeting

**Schedule**:
- Type: One-time
- Date: 2025-11-20
- Time: 2:00 PM
- Template: Firefox
- Timezone: America/Los_Angeles
- Pre-warm: 15 minutes
- Auto-terminate: Disable (you'll close it manually)

---

## Troubleshooting

### Session Didn't Start

**Check**:
1. Schedule is enabled (not paused)
2. Next run time hasn't passed yet
3. Timezone is correct
4. No quota limits exceeded

**Verify**:
- Go to **Scheduling** → **Scheduled Sessions**
- Check "Last Run" status
- If shows error, click for details

### Session Started Late

**Possible Causes**:
- Cluster resource constraints
- Pre-warming disabled
- Heavy cluster load

**Solutions**:
- Enable pre-warming
- Contact admin about resource availability

### Calendar Not Syncing

**Google Calendar**:
1. Check connection status
2. Click **Sync Now**
3. If fails, disconnect and reconnect
4. Check Google Calendar permissions

**Outlook Calendar**:
1. Verify Microsoft account is active
2. Re-authorize if token expired
3. Check Outlook sync settings

### Wrong Timezone

**Problem**: Sessions start at wrong time

**Solution**:
1. Delete schedule
2. Recreate with correct timezone
3. Use [Time Zone Converter](https://www.timeanddate.com/worldclock/converter.html)

---

## Advanced: Cron Expressions

Cron format: `minute hour day_of_month month day_of_week`

**Examples**:
- `0 9 * * *` - Every day at 9:00 AM
- `0 */4 * * *` - Every 4 hours
- `0 9 * * 1-5` - Weekdays at 9:00 AM
- `0 0 1 * *` - First of every month at midnight
- `0 12 * * 0` - Every Sunday at noon

**Testing**:
- Use [crontab.guru](https://crontab.guru) to validate
- Hover over schedule to see next 3 run times

---

## Best Practices

### ✅ Do

- Use descriptive schedule names
- Enable auto-termination to save resources
- Pre-warm sessions you need immediately
- Set realistic timezones
- Test one-time schedules before creating recurring ones

### ❌ Don't

- Create overlapping schedules for same template
- Forget to disable unused schedules
- Set very short auto-terminate durations
- Use aggressive pre-warming (wastes resources)
- Sync to multiple calendars (creates duplicates)

---

## FAQ

**Q: Can I have multiple schedules for the same template?**
A: Yes, but they must not overlap in time.

**Q: What happens if I'm using a session when auto-terminate triggers?**
A: Active sessions are not terminated - only idle sessions.

**Q: Can I manually start a scheduled session early?**
A: Yes, go to **Sessions** and create a new session with that template.

**Q: Do scheduled sessions count against my quota?**
A: Yes, they count toward your active session limit.

**Q: Can I schedule sessions for other users (admins)?**
A: No, users must create their own schedules. Admins can create templates.

**Q: What's the maximum auto-terminate duration?**
A: 24 hours (1440 minutes). Contact admin for longer durations.

---

**Need Help?**
- **Enterprise Features**: `/docs/ENTERPRISE_FEATURES.md`
- **Support**: support@streamspace.io

---

*Last Updated: 2025-11-15*
