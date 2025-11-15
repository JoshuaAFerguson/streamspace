# Multi-Factor Authentication Setup Guide

**Difficulty**: Beginner
**Time Required**: 5-10 minutes

Enhance your account security with multi-factor authentication (MFA).

---

## What is MFA?

Multi-factor authentication adds an extra layer of security by requiring a second form of verification beyond your password. StreamSpace supports three MFA methods:

- **📱 Authenticator App** (Recommended): Use Google Authenticator, Authy, or 1Password
- **💬 SMS**: Receive codes via text message
- **📧 Email**: Receive codes via email

---

## Setting Up Authenticator App MFA

### Step 1: Navigate to Security Settings

1. Login to StreamSpace
2. Click your avatar (top right)
3. Select **Security** from the menu

### Step 2: Choose Authenticator App

1. Click **Set Up** under "Authenticator App"
2. A QR code will appear

### Step 3: Scan QR Code

**Option A - Using Your Phone**:
1. Open your authenticator app (Google Authenticator, Authy, etc.)
2. Tap "Add Account" or "+"
3. Choose "Scan QR Code"
4. Point your camera at the QR code on screen

**Option B - Manual Entry**:
1. Open your authenticator app
2. Choose "Enter key manually"
3. Enter the secret key shown below the QR code
4. Set account name to "StreamSpace"

### Step 4: Verify Setup

1. Your authenticator app will display a 6-digit code
2. Enter this code in the verification field
3. Click **Verify**

### Step 5: Save Backup Codes

**⚠️ IMPORTANT**: Save these backup codes in a safe place!

1. Copy all 10 backup codes shown
2. Store them in a password manager or secure location
3. Each code can only be used once
4. Use these if you lose access to your phone

Click **Complete Setup** when done.

---

## Setting Up SMS MFA

### Step 1: Navigate to Security Settings

1. Login to StreamSpace
2. Click **Security** in the navigation menu

### Step 2: Configure Phone Number

1. Click **Set Up** under "SMS"
2. Enter your mobile phone number
3. Select your country code
4. Click **Send Code**

### Step 3: Verify Phone Number

1. Check your phone for a text message
2. Enter the 6-digit code received
3. Click **Verify**

### Step 4: Save Backup Codes

Same as authenticator app setup - save the 10 backup codes!

---

## Setting Up Email MFA

### Step 1: Navigate to Security Settings

1. Login to StreamSpace
2. Click **Security** in the navigation menu

### Step 2: Verify Email

1. Click **Set Up** under "Email"
2. Confirm your email address is correct
3. Click **Send Code**

### Step 3: Check Your Email

1. Open your email inbox
2. Look for email from "StreamSpace Security"
3. Copy the 6-digit code

### Step 4: Complete Setup

1. Enter the code in the verification field
2. Click **Verify**
3. Save your backup codes

---

## Using MFA at Login

After setup, every login will require your second factor:

1. Enter your username and password
2. You'll be prompted for your MFA code
3. Options:
   - **Authenticator App**: Enter the current 6-digit code
   - **SMS**: Click "Send Code" and enter the code received
   - **Email**: Click "Send Code" and check your email

---

## Using Backup Codes

If you lose access to your MFA device:

1. At the MFA prompt, click **Use backup code**
2. Enter one of your saved backup codes
3. Click **Verify**
4. ⚠️ That code is now invalid - you have 9 remaining

**Lost all backup codes?** Contact your administrator for account recovery.

---

## Managing Your MFA Methods

### View Active Methods

Navigate to **Security** → **Active MFA Methods** to see:
- Which methods are enabled
- Primary method (used first at login)
- Last time each method was used

### Remove an MFA Method

1. Go to **Security** → **Active MFA Methods**
2. Click the trash icon next to the method
3. Confirm removal
4. ⚠️ You must have at least one MFA method active

### Change Primary Method

1. Go to **Security** → **Active MFA Methods**
2. Click **Set as Primary** on your preferred method

---

## Troubleshooting

### Authenticator App Shows Wrong Code

**Problem**: Code is always invalid

**Solutions**:
1. Check phone time is correct (Settings → Date & Time → Auto)
2. Ensure time zone is correct
3. Try the next code that appears (codes refresh every 30 seconds)

### Not Receiving SMS Codes

**Problem**: No text message arrives

**Solutions**:
1. Check phone signal strength
2. Verify phone number is correct in settings
3. Wait 2-3 minutes (delays can occur)
4. Try **Send Code** again (once per minute)

### Not Receiving Email Codes

**Problem**: No email arrives

**Solutions**:
1. Check spam/junk folder
2. Verify email address in profile settings
3. Add `security@streamspace.io` to contacts
4. Wait 5 minutes before requesting a new code

### Lost Access to All MFA Methods

**Problem**: Can't login and don't have backup codes

**Solutions**:
1. Contact your administrator immediately
2. Provide:
   - Your username
   - Recent login times/locations
   - Reason for MFA access loss
3. Administrator can disable MFA for recovery

---

## Security Best Practices

### ✅ Do

- Use an authenticator app (most secure)
- Save backup codes in a password manager
- Enable MFA on your email account too
- Update your phone number if it changes
- Keep your recovery methods current

### ❌ Don't

- Share MFA codes with anyone (including staff)
- Save backup codes in plain text files
- Use the same phone for MFA and password recovery
- Disable MFA unless absolutely necessary
- Take screenshots of QR codes (security risk)

---

## FAQ

**Q: Can I use multiple MFA methods?**
A: Yes! You can enable all three methods and choose at login.

**Q: Which MFA method is most secure?**
A: Authenticator app is most secure, followed by SMS, then email.

**Q: Do I need MFA for every login?**
A: Yes, unless you check "Trust this device for 30 days" at login.

**Q: Can I use the same authenticator app for multiple accounts?**
A: Yes, each account shows as a separate entry with its own codes.

**Q: What happens if I get a new phone?**
A: Set up MFA again on the new device using the same steps. You can disable the old MFA method after verifying the new one works.

---

**Need Help?**
- **User Guide**: `/docs/ENTERPRISE_FEATURES.md`
- **Support**: support@streamspace.io
- **Emergency**: Contact your administrator

---

*Last Updated: 2025-11-15*
