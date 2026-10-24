# Networking — Real-World Catalog

Protocols, appliances, and concepts engineers actually design with. Goes
below cloud abstractions (BGP, MPLS, SD-WAN, zero trust, overlays, eBPF).
Icons here aren't provider-specific — they represent logical primitives
that show up in network diagrams alongside cloud services.

## Routing protocols

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.bgp_ebgp` | eBGP | routing | A | External BGP — between autonomous systems; full mesh or route servers. | ebgp, bgp | Two-AS peering over IXP |
| `net.bgp_ibgp` | iBGP | routing | A | Internal BGP — within one AS; full mesh or route reflectors. | ibgp | Route reflectors in DC |
| `net.bgp_route_reflector` | Route Reflector | routing | A | iBGP scaling without full mesh. | route reflector, rr | RR in spine |
| `net.bgp_confederation` | BGP Confederation | routing | E | Split a large AS into sub-ASes internally. | bgp confederation | ISP backbone segmentation |
| `net.bgp_community` | BGP Community | routing | A | Tag prefixes for policy (Well-known + RFC 7153). | bgp community | No-export on leak |
| `net.bgp_flowspec` | BGP FlowSpec | security | E | RFC 5575 — distribute L3/L4 filters via BGP. | flowspec | Upstream DDoS mitigation |
| `net.ospf` | OSPF | routing | A | Link-state IGP (areas, ABR, ASBR, LSAs). | ospf | DC backbone |
| `net.isis` | IS-IS | routing | A | Link-state used in SP + hyperscaler underlays. | is-is, isis | ISP underlay |
| `net.eigrp` | EIGRP | routing | A | Cisco hybrid distance-vector (DUAL). | eigrp | Enterprise Cisco net |
| `net.rip` | RIP | routing | E | Distance-vector (⚠ deprecated for modern use). | rip | Lab / tiny topology |
| `net.bfd` | BFD | routing | A | Sub-second failure detection under routing protocols. | bfd | BGP+BFD |

## Overlays & transport

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.mpls` | MPLS | transport | A | Label-switched paths; supports VPN / TE / L3VPN / EVPN-VXLAN. | mpls | WAN MPLS cloud |
| `net.mpls_ldp` | LDP | transport | A | Label Distribution Protocol for MPLS. | ldp | LSP setup |
| `net.mpls_rsvp_te` | RSVP-TE | transport | A | Traffic engineering LSPs w/ QoS. | rsvp-te | SP TE tunnels |
| `net.mpls_l3vpn` | MPLS L3VPN | transport | A | Customer VRF isolation over shared MPLS core (RFC 4364). | l3vpn, mpls vpn | Enterprise WAN |
| `net.evpn` | EVPN | overlay | A | MAC + IP learning in BGP; used with VXLAN/MPLS. | evpn | DC fabric |
| `net.vxlan` | VXLAN | overlay | A | UDP-encapped L2-in-L3 (RFC 7348). | vxlan | DC leaf-spine |
| `net.geneve` | GENEVE | overlay | A | Generic encapsulation; extensible TLVs; used by AWS GWLB. | geneve | AWS Gateway LB |
| `net.nvgre` | NVGRE | overlay | E | Microsoft L2-in-GRE overlay (niche). | nvgre | Legacy Hyper-V |
| `net.gre` | GRE | overlay | A | RFC 2890 generic encapsulation (often w/ IPsec). | gre | GRE-over-IPsec VPN |
| `net.ipinip` | IP-in-IP | overlay | A | RFC 2003 simple tunnel (used by Calico). | ip in ip | Calico overlay |
| `net.segment_routing` | Segment Routing | transport | E | SR-MPLS / SRv6 — source-routed steering. | sr, srv6 | SP TE + slicing |

## Switching & L2

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.vlan` | VLAN | switching | I | 802.1Q tagging — 4094 max per trunk. | vlan | VLAN per tenant |
| `net.qinq` | Q-in-Q | switching | A | 802.1ad — double tagging (S-tag + C-tag). | qinq, s-tag | SP metro aggregation |
| `net.stp` | STP / RSTP / MSTP | switching | A | Spanning tree (802.1D / 1w / 1s). | stp, spanning tree | Loop prevention |
| `net.lacp` | LACP | switching | A | 802.3ad link aggregation. | lacp, port channel, bond | Two-link LAG |
| `net.mlag` | MLAG / VPC | switching | A | Multi-chassis LAG across switch pair. | mlag, vpc, esi-lag | Dual-home server |
| `net.spine_leaf` | Spine-Leaf | fabric | A | CLOS fabric: every leaf 1-hop from every spine. | spine leaf, clos fabric | Modern DC |
| `net.3tier_legacy` | 3-Tier DC | fabric | I | Legacy core/agg/access (⚠ dated). | 3-tier, core agg access | Brownfield DC |

## NAT, addressing

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.nat_snat` | SNAT | nat | I | Source NAT — rewrite source IP on egress. | snat, source nat | Egress NAT |
| `net.nat_dnat` | DNAT | nat | I | Destination NAT — rewrite destination IP on ingress. | dnat, destination nat | Port forward |
| `net.nat_pat` | PAT | nat | I | Port Address Translation (multiple internal → one external IP). | pat, masquerade | Router overload |
| `net.nat_cgn` | Carrier-Grade NAT | nat | A | NAT444 at ISP scale (RFC 6598 100.64.0.0/10). | cgn, cgnat | ISP sharing IPv4 |
| `net.nat_64` | NAT64 | nat | A | IPv6-only clients reach IPv4 servers (RFC 6146). | nat64 | IPv6 islands |
| `net.rfc1918` | RFC 1918 | addressing | B | Private IPv4 blocks (10/8, 172.16/12, 192.168/16). | rfc 1918, private ip | 10.0.0.0/8 design |
| `net.ipv6` | IPv6 | addressing | I | 128-bit; SLAAC, ULA, link-local, global unicast. | ipv6, ip v6 | Dual-stack deployment |
| `net.dual_stack` | Dual-Stack | addressing | I | IPv4 + IPv6 simultaneously. | dual stack | Modern access net |
| `net.464xlat` | 464XLAT | addressing | E | IPv4-in-IPv6 on the device side (mobile). | 464xlat | Mobile IPv6-only |

## DNS

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.dns_recursive_resolver` | Recursive Resolver | dns | I | Caching resolver traversing DNS hierarchy for clients. | recursive resolver, dns cache | Corporate resolver |
| `net.dns_authoritative` | Authoritative DNS | dns | I | Answers for a zone it owns. | authoritative dns, ns | Zone for example.com |
| `net.dns_root` | DNS Root | dns | A | 13 logical root servers (A–M). | dns root | Lookup start |
| `net.dns_tld` | TLD | dns | A | `.com`, `.io`, etc. delegated by IANA. | tld | Apex record |
| `net.dns_dnssec` | DNSSEC | dns | A | Signed zones w/ RRSIG / DS / DNSKEY. | dnssec | Signed zone |
| `net.dns_dot` | DNS-over-TLS | dns | A | Port 853 TLS-encrypted DNS (RFC 7858). | dot, dns over tls | Encrypted resolver |
| `net.dns_doh` | DNS-over-HTTPS | dns | A | Port 443 HTTPS DNS (RFC 8484). | doh, dns over https | Browser DoH |
| `net.dns_a_record` | A Record | dns | B | IPv4 address for a hostname. | a record | api.example.com A |
| `net.dns_aaaa_record` | AAAA | dns | B | IPv6 address record. | aaaa record | IPv6 AAAA |
| `net.dns_cname` | CNAME | dns | B | Alias record. | cname | www → apex |
| `net.dns_mx` | MX | dns | B | Mail exchanger. | mx record | Gmail MX |
| `net.dns_srv` | SRV | dns | I | Service locator (port + priority). | srv record | Kerberos SRV |
| `net.dns_txt` | TXT | dns | I | Arbitrary strings (SPF, DKIM, ACME). | txt record | SPF record |
| `net.dns_caa` | CAA | dns | A | Which CAs may issue for the domain. | caa | Restrict LetsEncrypt |
| `net.split_horizon_dns` | Split-Horizon DNS | dns | A | Different answers for internal vs external resolvers. | split horizon dns | Private zone |
| `net.geo_dns` | GeoDNS | dns | A | Answer based on query source location. | geo dns | Closest-region routing |

## Firewall & security

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.firewall_stateful` | Stateful Firewall | security | I | Tracks connection state (conntrack). | stateful firewall | iptables --state |
| `net.firewall_stateless` | Stateless Firewall | security | I | Packet-by-packet ACL. | stateless firewall | AWS NACL |
| `net.ngfw` | NGFW | security | A | Next-gen firewall — app + user awareness, IDS, TLS decrypt. | ngfw, palo alto, fortinet | Palo Alto PA-5420 |
| `net.utm` | UTM | security | A | Unified threat mgmt — NGFW + AV + spam + content filter. | utm | SMB edge |
| `net.ids` | IDS | security | A | Passive intrusion detection (Snort, Suricata, Zeek). | ids, intrusion detection | Out-of-band sensor |
| `net.ips` | IPS | security | A | Inline prevention (drops / resets). | ips, intrusion prevention | Inline NGFW |
| `net.waf` | WAF | security | A | L7 inspection (OWASP CRS, bot mitigation). | waf, web app firewall | OWASP CRS 3.3 |
| `net.ddos_scrubber` | DDoS Scrubbing | security | A | Anycast scrubbing centers (Arbor, Radware, Cloudflare Magic). | ddos scrubber | Always-on scrubbing |
| `net.sase` | SASE | security | A | Secure Access Service Edge — ZTNA + SWG + CASB + FWaaS + SD-WAN in cloud. | sase | Zscaler / Prisma Access |
| `net.sse` | SSE | security | A | Security Service Edge — SASE minus SD-WAN. | sse, security service edge | Cloud SWG + CASB |
| `net.casb` | CASB | security | A | Cloud Access Security Broker — SaaS visibility + DLP. | casb | Netskope CASB |
| `net.swg` | Secure Web Gateway | security | A | URL filtering + TLS inspect + CASB. | swg, secure web gateway | Corporate SWG |
| `net.ztna` | ZTNA | security | A | Zero-Trust Network Access — identity-based app access. | ztna, beyondcorp | ZTNA replaces VPN |
| `net.dlp` | DLP | security | A | Data Loss Prevention — inline content inspection. | dlp | Outbound PII block |

## VPN

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.vpn_ipsec` | IPsec VPN | vpn | I | RFC 6071 — AH + ESP; tunnel + transport modes. | ipsec, ipsec vpn | Site-to-site VPN |
| `net.vpn_ikev1` | IKEv1 | vpn | A | Legacy key exchange (⚠ prefer IKEv2). | ikev1 | Legacy gear |
| `net.vpn_ikev2` | IKEv2 | vpn | A | RFC 7296 key exchange; MOBIKE for mobility. | ikev2 | Modern IPsec |
| `net.vpn_openvpn` | OpenVPN | vpn | I | SSL VPN using OpenSSL, TCP or UDP. | openvpn | Client VPN |
| `net.vpn_wireguard` | WireGuard | vpn | A | Lean kernel VPN w/ Noise framework. | wireguard, wg | Modern VPN |
| `net.vpn_l2tp` | L2TP | vpn | A | L2 tunnel (often with IPsec). | l2tp | Legacy client VPN |
| `net.vpn_sstp` | SSTP | vpn | E | Microsoft SSL VPN. | sstp | Windows VPN |
| `net.vpn_pptp` | PPTP | vpn | E | ⚠ Deprecated — insecure. | pptp | Legacy |
| `net.vpn_s2s` | Site-to-Site VPN | vpn | I | Gateway-to-gateway tunnel; BGP or static. | site to site vpn, s2s vpn | Branch ↔ HQ |
| `net.vpn_client_p2s` | Client VPN | vpn | I | Point-to-site (user device → gateway). | client vpn, point to site | Remote worker |

## Load balancing & proxies

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.lb_round_robin` | Round Robin | lb | B | Each target in sequence. | round robin | Basic DNS RR |
| `net.lb_least_connections` | Least Connections | lb | I | Route to backend with fewest active conns. | least connections | Long-lived sessions |
| `net.lb_weighted` | Weighted | lb | I | Weighted distribution (canary, large-vs-small). | weighted | 95/5 canary |
| `net.lb_ip_hash` | IP Hash | lb | I | Hash source IP → target; stickiness. | ip hash | Sticky sessions |
| `net.lb_consistent_hash` | Consistent Hash | lb | A | Ring-hash for cache locality. | consistent hash | CDN origin |
| `net.lb_maglev` | Maglev | lb | E | Google-style consistent hash (minimal disruption). | maglev | Envoy Maglev |
| `net.lb_least_response_time` | Least Response Time | lb | A | Route to lowest p50 backend. | least response time | HAProxy leastresponse |
| `net.reverse_proxy_nginx` | NGINX | proxy | I | Reverse proxy / static server / LB. | nginx | Fronting upstream |
| `net.reverse_proxy_haproxy` | HAProxy | proxy | A | L4/L7 reverse proxy + LB, ACL-heavy. | haproxy | Enterprise HA proxy |
| `net.reverse_proxy_envoy` | Envoy | proxy | A | xDS-configured proxy; basis for Istio/AGC/AWS App Mesh. | envoy | Sidecar mesh proxy |
| `net.reverse_proxy_traefik` | Traefik | proxy | I | Auto-discovery proxy; Docker/K8s labels. | traefik | Docker-Compose proxy |
| `net.reverse_proxy_caddy` | Caddy | proxy | I | Auto-HTTPS reverse proxy. | caddy | Auto Let's Encrypt |
| `net.forward_proxy` | Forward Proxy | proxy | I | Client-side outbound proxy for filtering/caching. | forward proxy | Corporate egress |
| `net.gslb` | GSLB | lb | A | Global Server LB via DNS or anycast. | gslb, global load balancer | F5 BIG-IP GTM |
| `net.anycast` | Anycast | network | A | Same IP advertised from many locations; BGP routes to nearest. | anycast | 1.1.1.1 |

## SD-WAN & WAN

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.sdwan_viptela` | Cisco Viptela | sdwan | A | Cisco SD-WAN (SD-WAN for ISR/XE SD-WAN + cEdge). | viptela, cisco sd-wan | Branch SD-WAN |
| `net.sdwan_versa` | Versa SASE | sdwan | A | Versa-branch cPE + cloud gateway. | versa | Enterprise SD-WAN |
| `net.sdwan_velocloud` | VMware VeloCloud | sdwan | A | VMware SD-WAN orchestrator + edge. | velocloud, vmware sd-wan | SD-WAN gateway |
| `net.sdwan_silverpeak` | Aruba Silver Peak | sdwan | A | EdgeConnect SD-WAN. | silverpeak, aruba sd-wan | SD-WAN edge |
| `net.sdwan_palo_prisma` | Prisma Access | sdwan | A | Palo Alto SASE + SD-WAN bundle. | prisma access | SASE rollout |
| `net.mpls_migration` | MPLS → SD-WAN | sdwan | A | Common migration pattern (cost-down, internet underlay). | mpls migration | SD-WAN over DIA |

## Observability / taps

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.span` | SPAN | observability | I | Local port mirror (switch-scoped). | span, port mirror | Mirror to IDS |
| `net.rspan` | RSPAN | observability | A | Remote SPAN over VLAN. | rspan | Cross-switch mirror |
| `net.erspan` | ERSPAN | observability | A | Encapsulated remote SPAN over GRE/IP. | erspan | Mirror across fabric |
| `net.netflow` | NetFlow | observability | A | Cisco flow export (5/9/IPFIX). | netflow, flow export | NetFlow collector |
| `net.sflow` | sFlow | observability | A | Sampled flow exports. | sflow | Data-plane sampling |
| `net.ipfix` | IPFIX | observability | A | IETF flow format (RFC 7011). | ipfix | Flow analytics |
| `net.tap` | Network Tap | observability | I | Passive physical tap device. | tap, network tap | Packet broker |
| `net.packet_broker` | Packet Broker | observability | A | Filters + load-balances mirrored traffic to tools. | packet broker | Tool farm aggregator |
| `net.ebpf` | eBPF | kernel | A | In-kernel programs for packet + syscall telemetry. | ebpf, bpf | Cilium, Pixie, Hubble |
| `net.xdp` | XDP | kernel | A | eBPF hook at NIC driver — DDoS drop, LB. | xdp | XDP-based LB |

## QoS / shaping

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `net.qos_dscp` | DSCP | qos | A | Differentiated Services Code Point (6 bits) for per-hop behavior. | dscp | Expedited Forwarding |
| `net.qos_wred` | WRED | qos | A | Weighted Random Early Detection — selective drop. | wred | Congestion avoidance |
| `net.qos_shaping` | Traffic Shaping | qos | A | Buffer + rate-limit to contract. | shaping, policer | Tiered WAN QoS |

---

**Entry count: ~80.** Covers wire-level routing + security + overlays expected in serious network diagrams.
