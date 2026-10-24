# Edge / IoT / Hybrid / Space — Component Catalog

Surface that touches the physical world — CDN PoPs, edge runtimes, 5G MEC,
IoT wireless/protocols, satellite, hybrid infrastructure. Includes the
"I didn't know that existed" tier (carrier MEC, satellite ground segment).

## Edge compute / CDN edge

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `edge.cloudflare_workers` | Cloudflare Workers | edge_runtime | I | V8 isolates at 330+ PoPs; sub-ms cold start. | workers, cloudflare workers | Auth at edge |
| `edge.cloudflare_durable_objects` | Durable Objects | edge_runtime | A | Coordinator actors at edge (single-writer consistency). | durable objects, do | Per-room chat state |
| `edge.cloudflare_d1` | Cloudflare D1 | edge_runtime | A | SQLite-on-edge with global read replicas. | d1, cloudflare d1 | Edge DB |
| `edge.cloudflare_kv` | Workers KV | edge_runtime | I | Eventually-consistent edge KV. | workers kv | Config store |
| `edge.cloudflare_r2` | Cloudflare R2 | storage | I | S3-compatible egress-free object store. | r2, cloudflare r2 | Egress-free assets |
| `edge.fastly_compute` | Fastly Compute | edge_runtime | I | WASM-based edge runtime (formerly Compute@Edge). | fastly compute, compute at edge | WASM edge logic |
| `edge.deno_deploy` | Deno Deploy | edge_runtime | I | V8 isolates w/ Deno runtime. | deno deploy | Edge TypeScript |
| `edge.vercel_edge_functions` | Vercel Edge Functions | edge_runtime | I | Next.js middleware + edge routes. | vercel edge | Next.js edge |
| `edge.netlify_edge_functions` | Netlify Edge Functions | edge_runtime | I | Deno-based edge. | netlify edge | Netlify edge |
| `edge.akamai_edgeworkers` | Akamai EdgeWorkers | edge_runtime | A | JS at 4000+ Akamai PoPs. | edgeworkers, akamai edgeworkers | Akamai edge |
| `edge.bunny_cdn` | Bunny CDN | cdn | I | Pull-zone CDN with scriptable edge. | bunny cdn | Cheap CDN |
| `edge.stackpath` | StackPath | cdn | E | Edge platform (⚠ restructured). | stackpath | Edge platform |

## 5G MEC / carrier edge

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `edge.5g_mec` | 5G MEC | mec | E | Multi-access Edge Computing — compute inside 5G packet core. | 5g mec, mec, multi-access edge | AR at stadium |
| `edge.aws_wavelength` | AWS Wavelength | mec | E | AWS zones inside Verizon/Vodafone/KDDI/SKT 5G networks. | wavelength, aws wavelength | 5G AR app |
| `edge.azure_edge_zones` | Azure Edge Zones | mec | E | Azure regions in carrier 5G networks (⚠ limited). | azure edge zone, private mec | Carrier-hosted compute |
| `edge.gcp_distributed_cloud_edge` | GCP Distributed Cloud Edge | mec | E | GCP-managed edge hardware + 5G partners. | gdc edge | Retail edge |
| `edge.open_ran` | Open RAN | telco | E | O-RAN disaggregated radio (RU/DU/CU). | o-ran, open ran | 5G ORAN deploy |

## On-prem / hybrid hardware

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `edge.aws_outposts` | AWS Outposts | hybrid | A | Rack or server of AWS hardware in customer DC. | outposts, aws outposts | Bank Outposts rack |
| `edge.aws_local_zone` | AWS Local Zone | hybrid | A | Metro-proximate AWS extension. | local zone, aws local zone | Metro latency |
| `edge.azure_stack_hci` | Azure Stack HCI | hybrid | A | Hyperconverged on customer HW. | stack hci | Branch HCI |
| `edge.azure_stack_edge` | Azure Stack Edge | hybrid | A | Rugged appliance (GPU optional). | stack edge | Edge AI appliance |
| `edge.gcp_distributed_cloud_hosted` | GCP Distributed Cloud Hosted | hybrid | E | Partner-hosted GCP region. | gdc hosted | Carrier-hosted GCP |
| `edge.vmware_cloud_foundation` | VMware Cloud Foundation | hybrid | A | VCF stack — vSphere + NSX + vSAN + Aria. | vcf, vmware cloud foundation | Private cloud |

## IoT gateways + edge runtimes

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `iot.gateway_generic` | IoT Gateway | iot | I | Generic protocol translator / data aggregator at edge. | iot gateway | MQTT → cloud |
| `iot.greengrass` | AWS IoT Greengrass | iot | A | Edge runtime — Lambda, ML inference, Stream Manager (see aws.md). | greengrass | Fleet edge |
| `iot.iot_edge_azure` | Azure IoT Edge | iot | A | Docker-module runtime on edge. | azure iot edge | Module runtime |
| `iot.opcua_server` | OPC UA Server | industrial | A | IEC 62541 industrial interop protocol. | opc ua, opcua | PLC → cloud |
| `iot.nvidia_jetson` | NVIDIA Jetson | edge_hw | A | GPU-equipped SoC boards (Nano, Xavier, Orin) for edge AI. | jetson, nvidia jetson | Robot vision |
| `iot.coral_tpu` | Google Coral | edge_hw | A | Edge TPU modules. | coral, edge tpu | Image classify edge |

## IoT protocols

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `iot.mqtt` | MQTT | protocol | I | Pub/sub protocol over TCP, optimized for constrained devices. | mqtt | Device telemetry |
| `iot.amqp` | AMQP | protocol | I | Reliable message broker protocol. | amqp | Robust queueing |
| `iot.coap` | CoAP | protocol | A | REST-like for constrained devices (RFC 7252) over UDP. | coap | Low-power REST |
| `iot.http2` | HTTP/2 | protocol | B | IETF HTTP/2 (multiplexed streams). | http/2 | Modern APIs |
| `iot.http3_quic` | HTTP/3 (QUIC) | protocol | A | RFC 9114 over QUIC (UDP). | http/3, quic | HTTP/3 mobile |
| `iot.grpc` | gRPC | protocol | I | HTTP/2-based RPC w/ Protobuf. | grpc | Microservices RPC |

## IoT wireless

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `iot.lorawan` | LoRaWAN | wireless | A | Long-range / low-power WAN (unlicensed). | lorawan, lora | Agri sensor net |
| `iot.nb_iot` | NB-IoT | wireless | A | 3GPP licensed narrow-band cellular IoT. | nb-iot | Utility meter |
| `iot.lte_m` | LTE-M | wireless | A | LTE for machines (CAT-M1). | lte-m, cat-m | Fleet tracker |
| `iot.sigfox` | Sigfox | wireless | E | Proprietary LPWAN (⚠ limited coverage). | sigfox | Ultra-low-power |
| `iot.zigbee` | Zigbee | wireless | A | 802.15.4 mesh (Zigbee Alliance). | zigbee | Smart home mesh |
| `iot.zwave` | Z-Wave | wireless | A | Low-power mesh (Silicon Labs). | z-wave | Home automation |
| `iot.thread` | Thread | wireless | A | 802.15.4 IPv6 mesh (Thread Group). | thread | Smart-home IPv6 |
| `iot.matter` | Matter | wireless | A | IPv6 / Thread / Wi-Fi interop (CSA). | matter | Smart-home standard |
| `iot.ble` | Bluetooth LE | wireless | I | Short-range low-power (Bluetooth SIG). | ble, bluetooth le | Beacons + wearables |
| `iot.wifi6` | Wi-Fi 6 / 6E | wireless | I | 802.11ax w/ 6 GHz band in 6E. | wifi 6, 802.11ax | High-density Wi-Fi |
| `iot.wifi7` | Wi-Fi 7 | wireless | A | 802.11be — MLO, 320 MHz channels. | wifi 7, 802.11be | 2025+ Wi-Fi |

## Industrial (OT) protocols

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `iot.modbus` | Modbus | industrial | A | Legacy serial / TCP PLC protocol. | modbus | PLC register read |
| `iot.profinet` | PROFINET | industrial | A | Siemens Ethernet-based industrial. | profinet | Siemens factory |
| `iot.ethercat` | EtherCAT | industrial | A | Hard-real-time fieldbus on Ethernet. | ethercat | Robotics control |
| `iot.canbus` | CAN bus | industrial | A | Automotive in-vehicle bus. | canbus, can | Vehicle diag |
| `iot.profibus` | PROFIBUS | industrial | A | Older Siemens field bus. | profibus | Legacy plant |

## Digital twin / simulation

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `iot.digital_twin_azure` | Azure Digital Twins | twin | A | DTDL-modeled twin graph (see azure.md). | azure digital twins | Smart-building twin |
| `iot.digital_twin_aws` | AWS IoT TwinMaker | twin | A | 3D scenes + entity + telemetry. | twinmaker | Warehouse twin |
| `iot.nvidia_omniverse` | NVIDIA Omniverse | twin | E | USD-based simulation platform for industrial twins. | omniverse | Factory sim |

## Space / ground segment

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `edge.aws_ground_station` | AWS Ground Station | space | E | Managed satellite antenna service (see aws.md). | aws ground station | LEO telemetry |
| `edge.azure_orbital` | Azure Orbital | space | E | Microsoft-operated ground stations + ground network. | orbital, azure orbital | LEO contacts |
| `edge.gcp_distributed_cloud_air_gapped` | GCP Distributed Cloud Air-Gapped | space | E | Air-gapped sovereign GCP (intel-community). | gdc air gapped | Sovereign compute |
| `edge.ksat` | KSAT | space | E | Kongsberg satellite ground services (polar + teleport). | ksat | Polar orbit ops |
| `edge.leaf_space` | Leaf Space | space | E | Commercial GSaaS provider. | leaf space | LEO cubesat |
| `edge.atlas_space` | ATLAS Space Ops | space | E | Federated ground-station network. | atlas space | Multi-station relay |
| `edge.starlink` | Starlink | space | I | SpaceX LEO constellation broadband. | starlink | Remote-site backhaul |
| `edge.oneweb` | OneWeb | space | E | LEO constellation (Eutelsat-merged). | oneweb | Enterprise LEO |
| `edge.iridium` | Iridium | space | E | 66-sat LEO network (voice + L-band data). | iridium | Remote SBD |
| `edge.kuiper` | Project Kuiper | space | E | Amazon LEO constellation (launching). | kuiper | Amazon LEO |
| `edge.globalstar` | Globalstar | space | E | LEO constellation — partnered w/ Apple emergency SOS. | globalstar | iPhone SOS partner |
| `iot.leo_cubesat` | CubeSat | space | E | 10×10×10 cm standardized spacecraft. | cubesat | Observation cubesat |
| `iot.geo_satellite` | GEO Satellite | space | E | Geostationary 35,786 km altitude. | geo satellite | Broadcast + VSAT |
| `iot.leo_satellite` | LEO Satellite | space | E | Low-earth orbit (160–2000 km). | leo satellite | Starlink / OneWeb |
| `iot.meo_satellite` | MEO Satellite | space | E | Medium-earth orbit (e.g. O3b, GPS). | meo satellite | O3b Networks |

## Submarine + backhaul

| Canonical | Display | Sub | Lvl | Description | Aliases | Example |
|---|---|---|---|---|---|---|
| `edge.submarine_cable` | Submarine Cable | backbone | E | Trans-oceanic fiber (MAREA, 2Africa, Dunant). | submarine cable | Trans-Atlantic |
| `edge.terrestrial_fiber` | Terrestrial Fiber | backbone | I | Long-haul fiber carrier backbone. | terrestrial fiber | National backbone |
| `edge.ixp` | Internet Exchange Point | peering | A | Neutral peering facility (AMS-IX, DE-CIX, LINX, NL-IX). | ixp, peering exchange | BGP peer at IXP |

---

**Entry count: ~55.** Covers the physical and orbital surface of modern architectures.
