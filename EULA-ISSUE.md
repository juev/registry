# 🚨 EULA Issue and Solution

## Problem Description

When starting Nexus, there's an issue with EULA (End User License Agreement) that blocks Docker registry functionality on port 8082.

### Symptoms

```bash
[WARN] Repository 'docker-registry' may already exist or configuration error
[ERROR] Failed to create Docker repository
```

```bash
curl http://localhost:8082/v2/
# Response: You must accept the End User License Agreement (EULA)
```

## ✅ Solution (2 minutes)

### 1. Open Nexus in browser

```
http://localhost:8081
```

### 2. Login

- **Username:** `admin`
- **Password:** `admin123`

### 3. Complete Setup Wizard

- ✅ **Accept EULA** (required!)
- ✅ Configure Anonymous Access (recommended)
- ✅ Complete setup

### 4. Verify result

```bash
curl http://localhost:8082/v2/
# Should return: {}
```

## 🔧 Project Updates

### README.md

- Added clear EULA warnings
- Updated Troubleshooting section
- Added manual EULA acceptance instructions

### configure-nexus.sh

- Added automatic port 8082 availability check
- Clear instructions when EULA issue detected
- Colored output with warnings

## ⚡ Why this happens?

1. **API creates repository** - but it's blocked at Nexus level
2. **EULA blocks ports** - including Docker registry on 8082
3. **Automatic acceptance via API doesn't work** - web interface required
4. **Solution only via browser** - this is Sonatype's requirement

## 🎯 Result

After accepting EULA:

- ✅ Docker registry works on port 8082
- ✅ Go application can connect to Nexus  
- ✅ Functions `remote.Image()` and `remote.Get()` work correctly
- ✅ Can push/pull Docker images
- ✅ Experiment ready for full testing!

## 📚 Additional Information

See main [README.md](./README.md) file for complete setup and usage instructions.