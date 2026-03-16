# UniFi API Reference

> Source: [beezly/unifi-apis](https://github.com/beezly/unifi-apis) (OpenAPI 3.1.0 specs auto-extracted from UniFi controllers)
>
> - **Network API version**: 10.2.93
> - **Protect API version**: 7.0.88
>
> Base URL: `/integration` (relative to controller address)

---

## Table of Contents

- [Authentication](#authentication)
- [Pagination & Filtering](#pagination--filtering)
- [Error Handling](#error-handling)
- [Network API](#network-api)
  - [Application Info](#application-info)
  - [Sites](#sites)
  - [Devices](#devices)
  - [Clients](#clients)
  - [Networks](#networks)
  - [WiFi Broadcasts](#wifi-broadcasts)
  - [Hotspot Vouchers](#hotspot-vouchers)
  - [Firewall Zones](#firewall-zones)
  - [Firewall Policies](#firewall-policies)
  - [ACL Rules](#acl-rules)
  - [DNS Policies](#dns-policies)
  - [Traffic Matching Lists](#traffic-matching-lists)
  - [Switching](#switching)
  - [Supporting Resources](#supporting-resources)
- [Protect API](#protect-api)
  - [Application Info (Protect)](#application-info-protect)
  - [NVR](#nvr)
  - [Cameras](#cameras)
  - [Camera PTZ Control](#camera-ptz-control)
  - [Lights](#lights)
  - [Sensors](#sensors)
  - [Chimes](#chimes)
  - [Viewers](#viewers)
  - [Live Views](#live-views)
  - [Device Asset Files](#device-asset-files)
  - [Alarm Manager](#alarm-manager)
  - [WebSocket Subscriptions](#websocket-subscriptions)
- [Network API Schemas](#network-api-schemas)
- [Protect API Schemas](#protect-api-schemas)

---

## Authentication

The Network API uses **API Keys** for authentication. Generate API Keys in the Integrations section of your UniFi application. The key is passed as a header with each request.

The Protect API uses session/cookie-based authentication (login → receive cookie → use cookie for subsequent requests).

## Pagination & Filtering

### Pagination

Most list endpoints accept:

| Parameter | Type    | Description                      |
|-----------|---------|----------------------------------|
| `offset`  | integer | Number of items to skip          |
| `limit`   | integer | Maximum number of items to return|

Response pages include: `offset`, `limit`, `count`, `totalCount`, `data[]`.

### Filtering

Many `GET` and `DELETE` endpoints support a `filter` query parameter with structured, URL-safe syntax.

#### Expression Types

**Property expressions** — apply functions to a property:
```
id.eq(123)
name.isNotNull()
createdAt.in(2025-01-01, 2025-01-05)
```

**Compound expressions** — combine with logical operators:
```
and(name.isNull(), createdAt.gt(2025-01-01))
or(name.isNull(), expired.isNull(), expiresAt.isNull())
```

**Negation expressions**:
```
not(name.like('guest*'))
```

#### Property Types

| Type        | Example                                    | Notes                                             |
|-------------|--------------------------------------------|----------------------------------------------------|
| `STRING`    | `'Hello, ''World''!'`                      | Wrapped in single quotes; escape `'` with `''`     |
| `INTEGER`   | `123`                                      | Starts with a digit                                |
| `DECIMAL`   | `123.321`                                  | May include decimal point                          |
| `TIMESTAMP` | `2025-01-29`, `2025-01-29T12:39:11Z`       | ISO 8601                                           |
| `BOOLEAN`   | `true`, `false`                            |                                                    |
| `UUID`      | `550e8400-e29b-41d4-a716-446655440000`     | Standard UUID format                               |
| `SET(...)`  | `[1, 2, 3, 4, 5]`                         | A set of unique values                             |

#### Filter Functions

| Function          | Args  | Meaning                 | Types                                              |
|-------------------|-------|-------------------------|----------------------------------------------------|
| `isNull`          | 0     | is null                 | all                                                |
| `isNotNull`       | 0     | is not null             | all                                                |
| `eq`              | 1     | equals                  | STRING, INTEGER, DECIMAL, TIMESTAMP, BOOLEAN, UUID  |
| `ne`              | 1     | not equals              | STRING, INTEGER, DECIMAL, TIMESTAMP, BOOLEAN, UUID  |
| `gt`              | 1     | greater than            | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `ge`              | 1     | greater than or equals  | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `lt`              | 1     | less than               | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `le`              | 1     | less than or equals     | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `like`            | 1     | pattern match           | STRING                                             |
| `in`              | 1+    | one of                  | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `notIn`           | 1+    | not one of              | STRING, INTEGER, DECIMAL, TIMESTAMP, UUID           |
| `isEmpty`         | 0     | set is empty            | SET                                                |
| `contains`        | 1     | set contains            | SET                                                |
| `containsAny`     | 1+    | set contains any of     | SET                                                |
| `containsAll`     | 1+    | set contains all of     | SET                                                |
| `containsExactly` | 1+    | set contains exactly    | SET                                                |

**Pattern matching** (`like`): `.` = any single char, `*` = any chars, `\` = escape.

## Error Handling

Errors return a standard JSON object:

```json
{
  "statusCode": 401,
  "statusName": "UNAUTHORIZED",
  "code": "AUTH_EXPIRED",
  "message": "Session token has expired",
  "timestamp": "2025-01-29T12:39:11Z",
  "requestPath": "/v1/sites",
  "requestId": "abc-123"
}
```

---

## Network API

Base path: `/integration/v1`

### Application Info

| Method | Path       | Summary              | OperationId |
|--------|------------|----------------------|-------------|
| GET    | `/v1/info` | Get Application Info | `getInfo`   |

**Response**: `applicationVersion` (string).

---

### Sites

| Method | Path        | Summary           | OperationId          |
|--------|-------------|-------------------|----------------------|
| GET    | `/v1/sites` | List Local Sites  | `getSiteOverviewPage`|

**Query params**: `offset`, `limit`, `filter`

**Response**: Paginated list of site overviews.

---

### Devices

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/pending-devices` | List Devices Pending Adoption | `getPendingDevicePage` |
| GET | `/v1/sites/{siteId}/devices` | List Adopted Devices | `getAdoptedDeviceOverviewPage` |
| POST | `/v1/sites/{siteId}/devices` | Adopt Devices | `adoptDevice` |
| GET | `/v1/sites/{siteId}/devices/{deviceId}` | Get Adopted Device Details | `getAdoptedDeviceDetails` |
| DELETE | `/v1/sites/{siteId}/devices/{deviceId}` | Remove (Unadopt) Device | `removeDevice` |
| POST | `/v1/sites/{siteId}/devices/{deviceId}/actions` | Execute Device Action | `executeAdoptedDeviceAction` |
| POST | `/v1/sites/{siteId}/devices/{deviceId}/interfaces/ports/{portIdx}/actions` | Execute Port Action | `executePortAction` |
| GET | `/v1/sites/{siteId}/devices/{deviceId}/statistics/latest` | Get Latest Device Statistics | `getAdoptedDeviceLatestStatistics` |

#### Path Parameters

| Parameter  | Type    | Description          |
|------------|---------|----------------------|
| `siteId`   | string  | Site identifier      |
| `deviceId` | string  | Device identifier    |
| `portIdx`  | integer | Port index on device |

#### Adopt Device — Request Body

```json
{
  "macAddress": "aa:bb:cc:dd:ee:ff",
  "ignoreDeviceLimit": false
}
```

#### Device Action — Request Body

```json
{
  "action": "restart"
}
```

#### Port Action — Request Body

```json
{
  "action": "cycle"
}
```

#### Device Overview — Key Fields

| Field               | Type    | Description                                  |
|---------------------|---------|----------------------------------------------|
| `id`                | string  | Device ID                                    |
| `macAddress`        | string  | MAC address                                  |
| `ipAddress`         | string  | IP address                                   |
| `name`              | string  | Device name                                  |
| `model`             | string  | Hardware model                               |
| `state`             | string  | Current state                                |
| `supported`         | boolean | Whether device is supported                  |
| `firmwareVersion`   | string  | Current firmware version                     |
| `firmwareUpdatable` | boolean | Whether a firmware update is available       |
| `features`          | object  | Feature flags (switching, accessPoint, etc.) |
| `interfaces`        | object  | Physical interfaces (ports, radios)          |

#### Device Details — Additional Fields

| Field             | Type   | Description                        |
|-------------------|--------|------------------------------------|
| `adoptedAt`       | string | Adoption timestamp                 |
| `provisionedAt`   | string | Provisioning timestamp             |
| `configurationId` | string | Configuration ID                   |
| `uplink`          | object | Uplink interface (parent device)   |

---

### Clients

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/clients` | List Connected Clients | `getConnectedClientOverviewPage` |
| GET | `/v1/sites/{siteId}/clients/{clientId}` | Get Connected Client Details | `getConnectedClientDetails` |
| POST | `/v1/sites/{siteId}/clients/{clientId}/actions` | Execute Client Action | `executeConnectedClientAction` |

#### Client Action — Request Body

```json
{
  "action": "authorize"
}
```

Actions include `authorize` and `unauthorize` for guest clients.

#### Client Overview — Key Fields

| Field         | Type   | Description                                    |
|---------------|--------|------------------------------------------------|
| `type`        | string | Client type (wired, wireless, vpn, guest)      |
| `id`          | string | Client ID                                      |
| `name`        | string | Client name                                    |
| `connectedAt` | string | Connection timestamp                           |
| `ipAddress`   | string | IP address                                     |
| `access`      | object | Access information (authorized status, etc.)   |

---

### Networks

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/networks` | List Networks | `getNetworksOverviewPage` |
| POST | `/v1/sites/{siteId}/networks` | Create Network | `createNetwork` |
| GET | `/v1/sites/{siteId}/networks/{networkId}` | Get Network Details | `getNetworkDetails` |
| PUT | `/v1/sites/{siteId}/networks/{networkId}` | Update Network | `updateNetwork` |
| DELETE | `/v1/sites/{siteId}/networks/{networkId}` | Delete Network | `deleteNetwork` |
| GET | `/v1/sites/{siteId}/networks/{networkId}/references` | Get Network References | `getNetworkReferences` |

#### Delete Network — Query Parameters

| Parameter | Type    | Description                    |
|-----------|---------|--------------------------------|
| `force`   | boolean | Force delete even if in use    |

#### Create/Update Network — Required Fields

| Field        | Type    | Description                            |
|--------------|---------|----------------------------------------|
| `name`       | string  | Network name                           |
| `enabled`    | boolean | Whether network is enabled             |
| `management` | boolean | Whether this is a management network   |
| `vlanId`     | integer | VLAN ID                                |

#### Create/Update Network — Optional Fields

| Field          | Type   | Description                |
|----------------|--------|----------------------------|
| `dhcpGuarding` | object | DHCP guarding config       |

#### Network Details — Response includes

Gateway-managed networks include IPv4/IPv6 configuration, DHCP settings, NAT outbound IP configuration, cellular backup settings, internet access, isolation, and mDNS forwarding.

---

### WiFi Broadcasts

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/wifi/broadcasts` | List WiFi Broadcasts | `getWifiBroadcastPage` |
| POST | `/v1/sites/{siteId}/wifi/broadcasts` | Create WiFi Broadcast | `createWifiBroadcast` |
| GET | `/v1/sites/{siteId}/wifi/broadcasts/{wifiBroadcastId}` | Get WiFi Broadcast Details | `getWifiBroadcastDetails` |
| PUT | `/v1/sites/{siteId}/wifi/broadcasts/{wifiBroadcastId}` | Update WiFi Broadcast | `updateWifiBroadcast` |
| DELETE | `/v1/sites/{siteId}/wifi/broadcasts/{wifiBroadcastId}` | Delete WiFi Broadcast | `deleteWifiBroadcast` |

#### Delete WiFi Broadcast — Query Parameters

| Parameter | Type    | Description                    |
|-----------|---------|--------------------------------|
| `force`   | boolean | Force delete even if in use    |

---

### Hotspot Vouchers

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/hotspot/vouchers` | List Vouchers | `getVouchers` |
| POST | `/v1/sites/{siteId}/hotspot/vouchers` | Generate Vouchers | `createVouchers` |
| DELETE | `/v1/sites/{siteId}/hotspot/vouchers` | Delete Vouchers (bulk, filter required) | `deleteVouchers` |
| GET | `/v1/sites/{siteId}/hotspot/vouchers/{voucherId}` | Get Voucher Details | `getVoucher` |
| DELETE | `/v1/sites/{siteId}/hotspot/vouchers/{voucherId}` | Delete Voucher | `deleteVoucher` |

#### Create Vouchers — Request Body

| Field                  | Type    | Required | Description                          |
|------------------------|---------|----------|--------------------------------------|
| `name`                 | string  | yes      | Voucher name                         |
| `timeLimitMinutes`     | integer | yes      | Duration in minutes                  |
| `count`                | integer | no       | Number of vouchers to generate       |
| `authorizedGuestLimit` | integer | no       | Max concurrent guests per voucher    |
| `dataUsageLimitMBytes` | integer | no       | Data cap in MB                       |
| `rxRateLimitKbps`      | integer | no       | Download rate limit in Kbps          |
| `txRateLimitKbps`      | integer | no       | Upload rate limit in Kbps            |

#### Voucher Details — Key Fields

| Field                  | Type    | Description                         |
|------------------------|---------|-------------------------------------|
| `id`                   | string  | Voucher ID                          |
| `code`                 | string  | Voucher code                        |
| `name`                 | string  | Voucher name                        |
| `createdAt`            | string  | Creation timestamp                  |
| `activatedAt`          | string  | First use timestamp                 |
| `expiresAt`            | string  | Expiration timestamp                |
| `expired`              | boolean | Whether voucher has expired         |
| `authorizedGuestCount` | integer | Current active guest count          |

---

### Firewall Zones

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/firewall/zones` | List Firewall Zones | `getFirewallZones` |
| POST | `/v1/sites/{siteId}/firewall/zones` | Create Custom Firewall Zone | `createFirewallZone` |
| GET | `/v1/sites/{siteId}/firewall/zones/{firewallZoneId}` | Get Firewall Zone | `getFirewallZone` |
| PUT | `/v1/sites/{siteId}/firewall/zones/{firewallZoneId}` | Update Firewall Zone | `updateFirewallZone` |
| DELETE | `/v1/sites/{siteId}/firewall/zones/{firewallZoneId}` | Delete Custom Firewall Zone | `deleteFirewallZone` |

#### Create/Update Zone — Required Fields

| Field        | Type     | Description              |
|--------------|----------|--------------------------|
| `name`       | string   | Zone name                |
| `networkIds` | string[] | Network IDs in this zone |

---

### Firewall Policies

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/firewall/policies` | List Firewall Policies | `getFirewallPolicies` |
| POST | `/v1/sites/{siteId}/firewall/policies` | Create Firewall Policy | `createFirewallPolicy` |
| GET | `/v1/sites/{siteId}/firewall/policies/ordering` | Get Policy Ordering | `getFirewallPolicyOrdering` |
| PUT | `/v1/sites/{siteId}/firewall/policies/ordering` | Reorder Policies | `updateFirewallPolicyOrdering` |
| GET | `/v1/sites/{siteId}/firewall/policies/{firewallPolicyId}` | Get Firewall Policy | `getFirewallPolicy` |
| PUT | `/v1/sites/{siteId}/firewall/policies/{firewallPolicyId}` | Update Firewall Policy | `updateFirewallPolicy` |
| DELETE | `/v1/sites/{siteId}/firewall/policies/{firewallPolicyId}` | Delete Firewall Policy | `deleteFirewallPolicy` |
| PATCH | `/v1/sites/{siteId}/firewall/policies/{firewallPolicyId}` | Patch Firewall Policy | `patchFirewallPolicy` |

#### Policy Ordering — Query Parameters (required)

| Parameter                    | Type   | Description               |
|------------------------------|--------|---------------------------|
| `sourceFirewallZoneId`       | string | Source zone ID             |
| `destinationFirewallZoneId`  | string | Destination zone ID        |

#### Create/Update Policy — Required Fields

| Field              | Type    | Description                                |
|--------------------|---------|--------------------------------------------|
| `name`             | string  | Policy name                                |
| `enabled`          | boolean | Whether policy is active                   |
| `action`           | object  | Action (type: `ALLOW`, `BLOCK`, `DROP`)    |
| `source`           | object  | Source zone and traffic filter              |
| `destination`      | object  | Destination zone and traffic filter         |
| `ipProtocolScope`  | object  | IP version and protocol matching            |
| `loggingEnabled`   | boolean | Enable traffic logging                      |

#### Optional Policy Fields

| Field                   | Type    | Description                          |
|-------------------------|---------|--------------------------------------|
| `description`           | string  | Policy description                   |
| `connectionStateFilter` | object  | Match by connection state            |
| `ipsecFilter`           | object  | IPSec matching                       |
| `schedule`              | object  | Time-based schedule                  |

#### Traffic Filters (source/destination)

Traffic filters can match by:
- IP address (single, range, or subnet)
- MAC address
- Network IDs
- VPN server IDs
- Site-to-site VPN tunnel IDs
- Application or application category (DPI)
- Domain
- Port (single, range, or list)
- Region/country
- IPv6 interface identifier

---

### ACL Rules

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/acl-rules` | List ACL Rules | `getAclRulePage` |
| POST | `/v1/sites/{siteId}/acl-rules` | Create ACL Rule | `createAclRule` |
| GET | `/v1/sites/{siteId}/acl-rules/ordering` | Get ACL Rule Ordering | `getAclRuleOrdering` |
| PUT | `/v1/sites/{siteId}/acl-rules/ordering` | Reorder ACL Rules | `updateAclRuleOrdering` |
| GET | `/v1/sites/{siteId}/acl-rules/{aclRuleId}` | Get ACL Rule | `getAclRule` |
| PUT | `/v1/sites/{siteId}/acl-rules/{aclRuleId}` | Update ACL Rule | `updateAclRule` |
| DELETE | `/v1/sites/{siteId}/acl-rules/{aclRuleId}` | Delete ACL Rule | `deleteAclRule` |

#### Create/Update ACL Rule — Required Fields

| Field     | Type    | Description           |
|-----------|---------|-----------------------|
| `type`    | string  | Rule type             |
| `name`    | string  | Rule name             |
| `enabled` | boolean | Whether rule is active|
| `action`  | string  | Rule action           |

---

### DNS Policies

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/dns/policies` | List DNS Policies | `getDnsPolicyPage` |
| POST | `/v1/sites/{siteId}/dns/policies` | Create DNS Policy | `createDnsPolicy` |
| GET | `/v1/sites/{siteId}/dns/policies/{dnsPolicyId}` | Get DNS Policy | `getDnsPolicy` |
| PUT | `/v1/sites/{siteId}/dns/policies/{dnsPolicyId}` | Update DNS Policy | `updateDnsPolicy` |
| DELETE | `/v1/sites/{siteId}/dns/policies/{dnsPolicyId}` | Delete DNS Policy | `deleteDnsPolicy` |

#### DNS Policy Types

The policy `type` discriminator determines additional fields. Types include: A record, AAAA record, CNAME record, MX record, SRV record, TXT record, forward domain policy.

---

### Traffic Matching Lists

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/traffic-matching-lists` | List Traffic Matching Lists | `getTrafficMatchingLists` |
| POST | `/v1/sites/{siteId}/traffic-matching-lists` | Create Traffic Matching List | `createTrafficMatchingList` |
| GET | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Get Traffic Matching List | `getTrafficMatchingList` |
| PUT | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Update Traffic Matching List | `updateTrafficMatchingList` |
| DELETE | `/v1/sites/{siteId}/traffic-matching-lists/{id}` | Delete Traffic Matching List | `deleteTrafficMatchingList` |

#### Create/Update — Required Fields

| Field  | Type   | Description      |
|--------|--------|------------------|
| `type` | string | List type        |
| `name` | string | List name        |

---

### Switching

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/sites/{siteId}/switching/lags` | List LAGs | `getLagPage` |
| GET | `/v1/sites/{siteId}/switching/lags/{lagId}` | Get LAG Details | `getLag` |
| GET | `/v1/sites/{siteId}/switching/mc-lag-domains` | List MC-LAG Domains | `getMcLagDomainPage` |
| GET | `/v1/sites/{siteId}/switching/mc-lag-domains/{id}` | Get MC-LAG Domain | `getMcLagDomain` |
| GET | `/v1/sites/{siteId}/switching/switch-stacks` | List Switch Stacks | `getSwitchStackPage` |
| GET | `/v1/sites/{siteId}/switching/switch-stacks/{id}` | Get Switch Stack | `getSwitchStack` |

> Note: Switching endpoints are **read-only** in the current API version.

---

### Supporting Resources

Read-only reference data endpoints:

| Method | Path | Summary | OperationId |
|--------|------|---------|-------------|
| GET | `/v1/countries` | List Countries | `getCountries` |
| GET | `/v1/dpi/applications` | List DPI Applications | `getDpiApplications` |
| GET | `/v1/dpi/categories` | List DPI Application Categories | `getDpiApplicationCategories` |
| GET | `/v1/sites/{siteId}/device-tags` | List Device Tags | `getDeviceTagPage` |
| GET | `/v1/sites/{siteId}/radius/profiles` | List RADIUS Profiles | `getRadiusProfileOverviewPage` |
| GET | `/v1/sites/{siteId}/wans` | List WAN Interfaces | `getWansOverviewPage` |
| GET | `/v1/sites/{siteId}/vpn/servers` | List VPN Servers | `getVpnServerPage` |
| GET | `/v1/sites/{siteId}/vpn/site-to-site-tunnels` | List Site-to-Site VPN Tunnels | `getSiteToSiteVpnTunnelPage` |

---

## Protect API

Base path: `/integration/v1`

### Application Info (Protect)

| Method | Path            | Summary                     |
|--------|-----------------|-----------------------------|
| GET    | `/v1/meta/info` | Get Protect application info|

**Response**: `applicationVersion` (string).

---

### NVR

| Method | Path       | Summary         |
|--------|------------|-----------------|
| GET    | `/v1/nvrs` | Get NVR Details |

#### NVR — Key Fields

| Field              | Type   | Description                |
|--------------------|--------|----------------------------|
| `id`               | string | NVR identifier             |
| `modelKey`         | string | Model key                  |
| `name`             | string | NVR name                   |
| `doorbellSettings` | object | Doorbell message settings  |

---

### Cameras

| Method | Path | Summary | Description |
|--------|------|---------|-------------|
| GET | `/v1/cameras` | Get All Cameras | List all cameras with details |
| GET | `/v1/cameras/{id}` | Get Camera Details | Get a specific camera |
| PATCH | `/v1/cameras/{id}` | Patch Camera Settings | Update camera settings |
| GET | `/v1/cameras/{id}/snapshot` | Get Camera Snapshot | Returns JPEG image (binary) |
| POST | `/v1/cameras/{id}/disable-mic-permanently` | Permanently Disable Mic | **Irreversible** unless factory reset |
| GET | `/v1/cameras/{id}/rtsps-stream` | Get RTSPS Streams | Get existing stream URLs |
| POST | `/v1/cameras/{id}/rtsps-stream` | Create RTSPS Streams | Create streams for quality levels |
| DELETE | `/v1/cameras/{id}/rtsps-stream` | Delete RTSPS Stream | Remove streams (query: `qualities`) |
| POST | `/v1/cameras/{id}/talkback-session` | Create Talkback Session | Get talkback stream URL + audio config |

#### Snapshot — Query Parameters

| Parameter     | Type   | Description                                  |
|---------------|--------|----------------------------------------------|
| `highQuality` | string | Force 1080P or higher resolution snapshot    |

#### Camera — Key Fields

| Field                 | Type    | Description                             |
|-----------------------|---------|-----------------------------------------|
| `id`                  | string  | Camera identifier                       |
| `modelKey`            | string  | Model key (`camera`)                    |
| `state`               | string  | Connection state                        |
| `name`                | string  | Camera name                             |
| `mac`                 | string  | MAC address                             |
| `isMicEnabled`        | boolean | Microphone enabled                      |
| `osdSettings`         | object  | On-screen display settings              |
| `ledSettings`         | object  | LED settings (status, welcome, flood)   |
| `lcdMessage`          | object  | LCD message (for doorbells)             |
| `micVolume`           | number  | Mic volume (0–100)                      |
| `activePatrolSlot`    | number  | Active PTZ patrol slot (0–4 or null)    |
| `videoMode`           | string  | Current video mode                      |
| `hdrType`             | string  | HDR mode setting                        |
| `featureFlags`        | object  | Capabilities (HDR, mic, speaker, etc.)  |
| `smartDetectSettings` | object  | Smart detection object/audio types      |

#### Patchable Camera Fields

| Field         | Type   | Description                           |
|---------------|--------|---------------------------------------|
| `name`        | string | Camera name                           |
| `osdSettings` | object | On-screen display settings            |
| `ledSettings` | object | LED settings                          |
| `lcdMessage`  | object | LCD message                           |
| `micVolume`   | number | Mic volume (0–100)                    |
| `videoMode`   | string | Video mode                            |
| `hdrType`     | string | HDR mode                              |
| `smartDetectSettings` | object | Smart detect object/audio types |

#### RTSPS Stream Qualities

Create request body:
```json
{
  "qualities": ["high", "medium", "low"]
}
```

Response contains stream URLs per quality level: `high`, `medium`, `low`, `package`.

#### Talkback Session — Response

| Field           | Type    | Description          |
|-----------------|---------|----------------------|
| `url`           | string  | Talkback stream URL  |
| `codec`         | string  | Audio codec          |
| `samplingRate`  | integer | Audio sampling rate   |
| `bitsPerSample` | integer | Bits per sample       |

---

### Camera PTZ Control

| Method | Path | Summary |
|--------|------|---------|
| POST | `/v1/cameras/{id}/ptz/goto/{slot}` | Move to Preset (slot 0–4) |
| POST | `/v1/cameras/{id}/ptz/patrol/start/{slot}` | Start Patrol (slot 0–4) |
| POST | `/v1/cameras/{id}/ptz/patrol/stop` | Stop Active Patrol |

All PTZ endpoints return `204 No Content` on success.

---

### Lights

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/lights` | Get All Lights |
| GET | `/v1/lights/{id}` | Get Light Details |
| PATCH | `/v1/lights/{id}` | Patch Light Settings |

#### Light — Key Fields

| Field                  | Type    | Description                             |
|------------------------|---------|-----------------------------------------|
| `id`                   | string  | Light identifier                        |
| `modelKey`             | string  | Model key (`light`)                     |
| `state`                | string  | Connection state                        |
| `name`                 | string  | Light name                              |
| `mac`                  | string  | MAC address                             |
| `lightModeSettings`    | object  | When/how the light activates            |
| `lightDeviceSettings`  | object  | PIR duration/sensitivity, LED level     |
| `isDark`               | boolean | Is scene dark                           |
| `isLightOn`            | boolean | Is LED currently on                     |
| `isLightForceEnabled`  | boolean | Is LED force-enabled                    |
| `lastMotion`           | number  | Unix timestamp of last PIR motion       |
| `isPirMotionDetected`  | boolean | Is PIR currently detecting motion       |
| `camera`               | object  | Associated camera reference             |

---

### Sensors

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/sensors` | Get All Sensors |
| GET | `/v1/sensors/{id}` | Get Sensor Details |
| PATCH | `/v1/sensors/{id}` | Patch Sensor Settings |

#### Sensor — Key Fields

| Field                      | Type    | Description                                  |
|----------------------------|---------|----------------------------------------------|
| `id`                       | string  | Sensor identifier                            |
| `modelKey`                 | string  | Model key (`sensor`)                         |
| `state`                    | string  | Connection state                             |
| `name`                     | string  | Sensor name                                  |
| `mac`                      | string  | MAC address                                  |
| `mountType`                | string  | Mount type (door, window, garage, leak)      |
| `batteryStatus`            | object  | Battery percentage and low status            |
| `stats`                    | object  | Light, humidity, temperature readings         |
| `isOpened`                 | boolean | Door/window/garage is opened                 |
| `openStatusChangedAt`      | number  | Last open/close timestamp                    |
| `isMotionDetected`         | boolean | Currently detecting motion                   |
| `motionDetectedAt`         | number  | Last motion timestamp                        |
| `alarmTriggeredAt`         | number  | Smoke/CO alarm timestamp                     |
| `leakDetectedAt`           | number  | Water leak timestamp                         |
| `externalLeakDetectedAt`   | number  | External water leak timestamp                |
| `tamperingDetectedAt`      | number  | Tampering detected timestamp                 |

#### Patchable Sensor Settings

| Field               | Type   | Description                                   |
|---------------------|--------|-----------------------------------------------|
| `name`              | string | Sensor name                                   |
| `lightSettings`     | object | Ambient light sensor (enabled, thresholds)    |
| `humiditySettings`  | object | Humidity sensor (enabled, thresholds)          |
| `temperatureSettings`| object| Temperature sensor (enabled, thresholds)      |
| `motionSettings`    | object | Motion detection (enabled, sensitivity)       |
| `alarmSettings`     | object | Smoke/CO alarm (enabled)                      |
| `leakSettings`      | object | Leak detection (internal/external enabled)    |

---

### Chimes

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/chimes` | Get All Chimes |
| GET | `/v1/chimes/{id}` | Get Chime Details |
| PATCH | `/v1/chimes/{id}` | Patch Chime Settings |

#### Chime — Key Fields

| Field          | Type     | Description                       |
|----------------|----------|-----------------------------------|
| `id`           | string   | Chime identifier                  |
| `modelKey`     | string   | Model key (`chime`)               |
| `state`        | string   | Connection state                  |
| `name`         | string   | Chime name                        |
| `mac`          | string   | MAC address                       |
| `cameraIds`    | string[] | Associated camera IDs             |
| `ringSettings` | object[] | Ring configuration per camera     |

#### Ring Settings

| Field         | Type    | Description              |
|---------------|---------|--------------------------|
| `cameraId`    | string  | Associated camera ID     |
| `repeatTimes` | integer | Number of ring repeats   |
| `ringtoneId`  | string  | Ringtone identifier      |
| `volume`      | number  | Ring volume              |

---

### Viewers

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/viewers` | Get All Viewers |
| GET | `/v1/viewers/{id}` | Get Viewer Details |
| PATCH | `/v1/viewers/{id}` | Patch Viewer Settings |

#### Viewer — Key Fields

| Field         | Type   | Description                        |
|---------------|--------|------------------------------------|
| `id`          | string | Viewer identifier                  |
| `modelKey`    | string | Model key (`viewer`)               |
| `state`       | string | Connection state                   |
| `name`        | string | Viewer name                        |
| `mac`         | string | MAC address                        |
| `liveview`    | object | Associated live view configuration |
| `streamLimit` | number | Max parallel live streams          |

---

### Live Views

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/liveviews` | Get All Live Views |
| POST | `/v1/liveviews` | Create Live View |
| GET | `/v1/liveviews/{id}` | Get Live View Details |
| PATCH | `/v1/liveviews/{id}` | Patch Live View Configuration |

#### Live View — Key Fields

| Field      | Type     | Description                    |
|------------|----------|--------------------------------|
| `id`       | string   | Live view identifier           |
| `modelKey` | string   | Model key (`liveview`)         |
| `name`     | string   | Live view name                 |
| `isDefault`| boolean  | Is this the default view       |
| `isGlobal` | boolean  | Is this globally visible       |
| `owner`    | string   | Owner user ID                  |
| `layout`   | string   | Layout type                    |
| `slots`    | array    | Camera slot assignments        |

---

### Device Asset Files

| Method | Path | Summary |
|--------|------|---------|
| GET | `/v1/files/{fileType}` | List Device Asset Files |
| POST | `/v1/files/{fileType}` | Upload Device Asset File |

#### File Types

The `fileType` path parameter specifies the asset category. Supported MIME types for upload: `image/gif`, `image/jpeg`, `image/png`.

#### File Schema — Response

| Field          | Type   | Description              |
|----------------|--------|--------------------------|
| `name`         | string | Unique file ID           |
| `type`         | string | File type                |
| `originalName` | string | Original filename        |
| `path`         | string | Filesystem path          |

---

### Alarm Manager

| Method | Path | Summary |
|--------|------|---------|
| POST | `/v1/alarm-manager/webhook/{id}` | Send Webhook to Alarm Manager |

Triggers configured alarms. The `id` path parameter is a user-defined string that matches the alarm's configured trigger ID.

Returns `204 No Content` on success.

---

### WebSocket Subscriptions

| Endpoint | Summary |
|----------|---------|
| `GET /v1/subscribe/devices` | Device change events (add, update, remove) |
| `GET /v1/subscribe/events` | Protect events (motion, smart detect, ring, sensor, etc.) |

#### Device Events

Messages are one of: `deviceAdd`, `deviceUpdate`, `deviceRemove` — each wrapping the relevant device object.

#### Protect Event Types

| Event Type | Description |
|------------|-------------|
| `cameraMotionEvent` | Camera motion detection start/end |
| `cameraSmartDetectZoneEvent` | Smart video zone detection |
| `cameraSmartDetectLineEvent` | Smart video line detection |
| `cameraSmartDetectLoiterEvent` | Smart video loiter detection |
| `cameraSmartDetectAudioEvent` | Smart audio detection |
| `ringEvent` | Doorbell ring button pressed |
| `lightMotionEvent` | Floodlight motion detected |
| `sensorMotionEvent` | Sensor motion start/end |
| `sensorOpenEvent` | Door/window/garage opened |
| `sensorClosedEvent` | Door/window/garage closed |
| `sensorAlarmEvent` | Smoke/CO alarm triggered |
| `sensorWaterLeakEvent` | Water leak detected |
| `sensorExtremeValueEvent` | Sensor metric out of range |
| `sensorTamperEvent` | Sensor tampered with |
| `sensorBatteryLowEvent` | Battery level low |
| `sensorSmokeTestEvent` | Smoke detector test initiated |

---

## Network API Schemas

Key request/response schemas used across the Network API.

### Paginated Response (all list endpoints)

```json
{
  "offset": 0,
  "limit": 25,
  "count": 10,
  "totalCount": 42,
  "data": [...]
}
```

### Device Action Request

```json
{ "action": "restart" }
```

### Port Action Request

```json
{ "action": "cycle" }
```

### Client Action Request

```json
{ "action": "authorize" | "unauthorize" }
```

### Guest Access Authorization Request

```json
{
  "action": "authorize",
  "expiresAt": "2025-06-01T00:00:00Z",
  "dataUsageLimitMBytes": 1024,
  "rxRateLimitKbps": 5000,
  "txRateLimitKbps": 1000
}
```

### Create/Update Network

```json
{
  "name": "IoT Network",
  "enabled": true,
  "management": false,
  "vlanId": 30,
  "dhcpGuarding": {}
}
```

### Create/Update Firewall Policy

```json
{
  "name": "Block IoT to LAN",
  "enabled": true,
  "action": { "type": "BLOCK" },
  "source": {
    "zoneId": "<source-zone-id>",
    "trafficFilter": { "type": "..." }
  },
  "destination": {
    "zoneId": "<dest-zone-id>"
  },
  "ipProtocolScope": {
    "ipVersion": "DUAL_STACK"
  },
  "loggingEnabled": true
}
```

### Create/Update Firewall Zone

```json
{
  "name": "IoT Zone",
  "networkIds": ["<network-id-1>", "<network-id-2>"]
}
```

### Hotspot Voucher Creation

```json
{
  "name": "Guest Day Pass",
  "count": 10,
  "timeLimitMinutes": 1440,
  "authorizedGuestLimit": 1,
  "dataUsageLimitMBytes": 2048,
  "rxRateLimitKbps": 10000,
  "txRateLimitKbps": 5000
}
```

### Device Adoption

```json
{
  "macAddress": "aa:bb:cc:dd:ee:ff",
  "ignoreDeviceLimit": false
}
```

---

## Protect API Schemas

Key request/response schemas used across the Protect API.

### Generic Error

```json
{
  "error": "NotFoundError",
  "name": "Camera not found",
  "cause": "No camera with the given ID exists"
}
```

### OSD Settings

```json
{
  "isNameEnabled": true,
  "isDateEnabled": true,
  "isLogoEnabled": false,
  "isDebugEnabled": false,
  "overlayLocation": "topLeft"
}
```

### LED Settings

```json
{
  "isEnabled": true,
  "welcomeLed": true,
  "floodLed": false
}
```

### Smart Detect Settings

```json
{
  "objectTypes": ["person", "vehicle", "animal"],
  "audioTypes": ["smoke_cmonx", "bark"]
}
```

### Light Mode Settings

```json
{
  "mode": "motion",
  "enableAt": { "type": "dark" }
}
```

### Sensor Threshold Settings (humidity/temperature/light)

```json
{
  "isEnabled": true,
  "margin": 2.0,
  "lowThreshold": 30.0,
  "highThreshold": 60.0
}
```

### Ring Settings (Chime)

```json
{
  "cameraId": "<camera-id>",
  "repeatTimes": 3,
  "ringtoneId": "default",
  "volume": 80
}
```

### Common Device Fields (all Protect devices)

All Protect device objects share:

| Field      | Type   | Description          |
|------------|--------|----------------------|
| `id`       | string | Unique identifier    |
| `modelKey` | string | Device type key      |
| `state`    | string | Connection state     |
| `name`     | string | Device name          |
| `mac`      | string | MAC address          |
