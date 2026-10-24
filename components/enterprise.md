# Enterprise Systems — Component Catalog

Third-party platforms and protocols that show up in real enterprise
architectures — the boxes a cloud diagram integrates with, not within.

## Identity providers

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.okta` | Okta | identity | I | Enterprise IdP + workforce + customer identity (Auth0). | okta | SSO + MFA |
| `ent.auth0` | Auth0 | identity | I | Developer-focused CIAM (Okta-owned). | auth0 | Consumer login |
| `ent.ping_identity` | Ping Identity | identity | A | Enterprise IdP + API Intelligence. | ping, ping identity | Regulated SSO |
| `ent.onelogin` | OneLogin | identity | I | IdP + PAM. | onelogin | SSO |
| `ent.jumpcloud` | JumpCloud | identity | I | Directory + device + SSO. | jumpcloud | Cloud AD alt |
| `ent.forgerock` | ForgeRock | identity | A | Enterprise IAM suite (Ping-owned). | forgerock | Gov-tier IAM |
| `ent.keycloak` | Keycloak | identity | A | Open-source IdP (Red Hat). | keycloak | Self-hosted IdP |
| `ent.authentik` | Authentik | identity | A | Open-source IdP. | authentik | Self-hosted SSO |
| `ent.active_directory` | Active Directory | identity | B | On-prem domain / forest / GPOs / trusts. | ad, active directory, ms ad | Legacy domain |
| `ent.ldap` | LDAP | identity | I | Directory access protocol (RFC 4511). | ldap | Legacy directory |
| `ent.kerberos` | Kerberos | identity | A | Ticket-granting auth (MIT). | kerberos | AD tickets |
| `ent.radius` | RADIUS | identity | A | Network auth protocol (VPN, Wi-Fi). | radius | WPA2-Enterprise |
| `ent.tacacs_plus` | TACACS+ | identity | A | Device admin auth (Cisco). | tacacs, tacacs+ | Router admin auth |

## Auth protocols

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.saml_2` | SAML 2.0 | auth | I | XML-based federation. | saml, saml 2.0 | SSO to SaaS |
| `ent.oauth_2_1` | OAuth 2.1 | auth | I | Authorization framework (RFC 6749 + new BCPs). | oauth, oauth2, oauth 2.1 | API delegation |
| `ent.oidc` | OpenID Connect | auth | I | Identity layer on OAuth 2. | oidc, openid connect | SSO via ID token |
| `ent.webauthn` | WebAuthn | auth | A | Passkeys (public-key credential). | webauthn, passkey | Passwordless login |
| `ent.fido2` | FIDO2 | auth | A | Hardware keys + WebAuthn. | fido2, yubikey | Security key |
| `ent.scim` | SCIM | identity | A | Cross-domain identity provisioning (RFC 7644). | scim | User provisioning |
| `ent.pkcs11` | PKCS#11 | crypto | A | HSM API for smart cards + crypto tokens. | pkcs11 | HSM integration |

## ERP

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.sap_s4hana` | SAP S/4HANA | erp | I | Next-gen SAP ERP on HANA. | s/4hana, sap s4 | ERP migration |
| `ent.sap_ecc` | SAP ECC | erp | A | Predecessor ERP (2027 mainstream support end). | sap ecc, ecc | Legacy SAP |
| `ent.sap_btp` | SAP BTP | integration | A | Business Technology Platform (extensions, integration, data, AI). | sap btp | SAP side-by-side extensions |
| `ent.sap_hana` | SAP HANA | database | A | In-memory columnar DB + app server. | sap hana | OLTP+OLAP DB |
| `ent.oracle_fusion` | Oracle Fusion | erp | I | Oracle Fusion Apps (ERP / HCM / SCM). | oracle fusion, fusion apps | Oracle ERP Cloud |
| `ent.oracle_ebs` | Oracle E-Business Suite | erp | A | Legacy on-prem ERP. | oracle ebs, ebs | Legacy EBS |
| `ent.oracle_netsuite` | NetSuite | erp | I | SMB-focused cloud ERP (Oracle). | netsuite | Small-biz ERP |
| `ent.workday` | Workday | hcm | I | HCM + Financials SaaS. | workday | HR system of record |
| `ent.microsoft_dynamics_365` | Dynamics 365 | erp | I | MS ERP + CRM suite. | dynamics 365, d365 | CRM + ERP |
| `ent.infor` | Infor | erp | A | Industry-specific ERP (mfg, healthcare). | infor | Industry ERP |

## CRM / sales / service

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.salesforce` | Salesforce | crm | B | Sales/Service/Marketing/Commerce/Data/Einstein/Tableau. | salesforce, sfdc | CRM |
| `ent.salesforce_data_cloud` | Salesforce Data Cloud | crm | A | Customer data platform (formerly CDP). | data cloud, salesforce cdp | Unified customer profile |
| `ent.hubspot` | HubSpot | crm | I | SMB-friendly CRM + marketing. | hubspot | Inbound marketing |
| `ent.zendesk` | Zendesk | service | I | Customer support platform. | zendesk | Ticket system |
| `ent.freshworks` | Freshworks | service | I | Fresh suite (CRM/ITSM/support). | freshworks, freshdesk | Support SaaS |

## ITSM / IT Ops

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.servicenow` | ServiceNow | itsm | I | ITSM + ITOM + HRSD + CSM + SecOps + GRC + Now Assist. | servicenow, snow | Ticketing + CMDB |
| `ent.jira_service_management` | Jira Service Management | itsm | I | Atlassian ITSM. | jsm, jira service | DevOps ITSM |
| `ent.pagerduty` | PagerDuty | incident | I | Incident response + on-call. | pagerduty | Page on-call |
| `ent.opsgenie` | Opsgenie | incident | I | On-call + escalation (Atlassian). | opsgenie | Atlassian paging |
| `ent.grafana_oncall` | Grafana OnCall | incident | A | Open-source on-call (paired w/ Grafana). | grafana oncall | OSS alternative |
| `ent.statuspage` | Statuspage | incident | I | Public status page (Atlassian). | statuspage | Customer-facing status |

## Observability / SIEM (SaaS)

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.datadog` | Datadog | observability | I | Metrics + logs + APM + RUM + security. | datadog, dd | SaaS observability |
| `ent.new_relic` | New Relic | observability | I | Observability SaaS. | new relic | APM |
| `ent.dynatrace` | Dynatrace | observability | A | Autodiscover APM + AI (Davis). | dynatrace | Enterprise APM |
| `ent.appdynamics` | AppDynamics | observability | A | Cisco APM. | appdynamics | Cisco APM |
| `ent.splunk` | Splunk | siem | A | SIEM + observability (Cisco-owned). | splunk, splunk es | SIEM |
| `ent.elastic_siem` | Elastic Security | siem | A | SIEM on ELK. | elastic siem, elk siem | Open SIEM |
| `ent.sumo_logic` | Sumo Logic | siem | A | Cloud-native SIEM + logs. | sumo logic | SaaS log SIEM |
| `ent.honeycomb` | Honeycomb | observability | A | Event-based observability (high-cardinality). | honeycomb | Wide-event observability |
| `ent.lightstep` | Lightstep | observability | A | Traces + SLOs (ServiceNow). | lightstep | Microservice traces |

## Security / EDR / PAM

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.crowdstrike` | CrowdStrike Falcon | security | A | EDR + threat intel + identity + cloud. | crowdstrike, falcon | Enterprise EDR |
| `ent.sentinelone` | SentinelOne | security | A | Autonomous EDR. | sentinelone, s1 | ML-driven EDR |
| `ent.carbon_black` | Carbon Black | security | A | VMware EDR. | carbon black, vmware edr | Legacy EDR |
| `ent.defender_for_endpoint` | Defender for Endpoint | security | I | Microsoft EDR (covered in azure). | mde | MS EDR |
| `ent.zscaler_zia` | Zscaler Internet Access | security | A | Cloud SWG + inline inspection. | zscaler, zia | Secure internet |
| `ent.zscaler_zpa` | Zscaler Private Access | security | A | ZTNA. | zpa | App-level access |
| `ent.netskope` | Netskope | security | A | SSE: CASB + SWG + ZTNA. | netskope | SSE platform |
| `ent.palo_prisma_access` | Palo Alto Prisma Access | security | A | SASE cloud (PAN-OS in cloud). | prisma access | SASE |
| `ent.cyberark` | CyberArk | pam | A | Enterprise PAM + secrets. | cyberark | Vault + PAM |
| `ent.hashicorp_vault` | HashiCorp Vault | secrets | I | Secrets + dynamic creds + PKI + transit encrypt. | vault, hashicorp vault | Dynamic DB creds |
| `ent.hashicorp_boundary` | HashiCorp Boundary | security | A | Identity-based session to hosts/apps. | boundary | ZTNA to SSH/DB |
| `ent.beyondtrust` | BeyondTrust | pam | A | PAM + remote support. | beyondtrust | PAM |
| `ent.tanium` | Tanium | endpoint | A | Endpoint mgmt + incident response. | tanium | Endpoint platform |

## Dev / SCM / registries

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.github` | GitHub | scm | B | Git hosting + Actions + Packages + Advanced Security. | github, gh | Hosted Git |
| `ent.github_actions` | GitHub Actions | ci | I | Event-driven CI/CD on GitHub. | github actions | CI workflow |
| `ent.github_enterprise_server` | GitHub Enterprise Server | scm | A | Self-hosted GitHub. | github enterprise, ghes | On-prem GitHub |
| `ent.gitlab` | GitLab | scm | I | Single app: Repos + CI + Registry + Security. | gitlab | Self-hosted DevOps |
| `ent.bitbucket` | Bitbucket | scm | I | Atlassian Git hosting. | bitbucket | Atlassian DevOps |
| `ent.gitea` | Gitea | scm | I | Lightweight self-hosted Git. | gitea | Small-team Git |
| `ent.artifactory` | Artifactory | registry | A | JFrog universal artifact manager. | artifactory, jfrog | Universal repo |
| `ent.nexus` | Nexus Repository | registry | A | Sonatype package manager. | nexus, sonatype | Maven + npm |
| `ent.harbor` | Harbor | registry | I | CNCF container registry. | harbor | Self-hosted OCI |

## Collaboration / messaging

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.slack` | Slack | collab | B | Team messaging (Salesforce). | slack | ChatOps |
| `ent.microsoft_teams` | Microsoft Teams | collab | B | Teams chat + meetings + channels. | teams, ms teams | Enterprise chat |
| `ent.discord` | Discord | collab | I | Community + gaming chat. | discord | Community |
| `ent.zoom` | Zoom | collab | B | Meetings + chat + phone + events. | zoom | Video meeting |
| `ent.google_workspace` | Google Workspace | productivity | B | Gmail / Drive / Docs / Meet. | workspace, gsuite | Productivity |
| `ent.microsoft_365` | Microsoft 365 | productivity | B | M365 — Outlook / Office / OneDrive / SharePoint / Teams. | m365, office 365 | Productivity |

## Backup / DR

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.veeam` | Veeam | backup | A | Backup + DR for VMware / Hyper-V / cloud. | veeam | VM backup |
| `ent.commvault` | Commvault | backup | A | Enterprise backup + DR. | commvault | Heterogeneous backup |
| `ent.rubrik` | Rubrik | backup | A | Cloud-ready data security + backup. | rubrik | Ransomware recovery |
| `ent.cohesity` | Cohesity | backup | A | Hyperconverged backup + DR. | cohesity | Unified backup |
| `ent.druva` | Druva | backup | A | SaaS-native backup. | druva | Endpoint + cloud backup |

## Payment / billing

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `ent.stripe` | Stripe | payments | I | Payments + billing + Terminal + Atlas. | stripe | Online payments |
| `ent.paystack` | Paystack | payments | I | African payments (Stripe-owned). | paystack | NGN payments |
| `ent.adyen` | Adyen | payments | A | Global unified commerce payments. | adyen | International |
| `ent.checkout` | Checkout.com | payments | A | Global payments. | checkout.com | Enterprise payments |
| `ent.paypal` | PayPal | payments | B | Consumer payments. | paypal | Buyer payment |

---

**Entry count: ~60.** Covers the enterprise SaaS a modern cloud architecture typically integrates with.
