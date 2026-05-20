import { useState, useEffect, useCallback, useMemo, useId, useRef } from 'react';
import { RefreshCw, AlertTriangle, CheckCircle, XCircle } from 'lucide-react';
import { PiholeManagementList } from '@/components/PiholeManagementList';
import { getConfig, patchConfig } from '@/lib/api/config';
import { usePiholes } from '@/providers/PiholeProvider';
import type { PiholeConfig, PatchConfigRequest, PatchConfigResponse } from '@/types/config';
import styles from './Settings.module.scss';

type Tab = 'nodes' | 'pihole-config';

// ---------- helpers ----------

function isDrifted(perNode: Record<string, PiholeConfig>, key: (c: PiholeConfig) => unknown): boolean {
	const vals = Object.values(perNode).map(key);
	if (vals.length < 2) return false;
	const first = JSON.stringify(vals[0]);
	return vals.some((v) => JSON.stringify(v) !== first);
}

// ---------- Config form state ----------

type ConfigFormState = {
	// DNS
	upstreams: string; // newline-separated
	interface: string;
	port: string;
	dnssec: boolean;
	domainNeeded: boolean;
	expandHosts: boolean;
	localTTL: string;
	blockingmode: string;
	blockingipv4: string;
	blockingipv6: string;
	rateLimitCount: string;
	rateLimitInterval: string;
	revServerActive: boolean;
	revServerCIDR: string;
	revServerTarget: string;
	revServerDomain: string;
	piholePTR: string;
	querylogEnabled: boolean;
	// Misc
	privacylevel: string;
	// FTL
	queryDisplay: string;
	maxDBdays: string;
	DBinterval: string;
	// Webserver
	excludeClients: string; // newline-separated
	excludeDomains: string;
	maxHistory: string;
	// Resolver
	resolveIPv4: boolean;
	resolveIPv6: boolean;
	networkNames: boolean;
};

function configToForm(cfg: PiholeConfig): ConfigFormState {
	return {
		upstreams: cfg.dns.upstreams.join('\n'),
		interface: cfg.dns.interface,
		port: String(cfg.dns.port),
		dnssec: cfg.dns.dnssec,
		domainNeeded: cfg.dns.domainNeeded,
		expandHosts: cfg.dns.expandHosts,
		localTTL: String(cfg.dns.localTTL),
		blockingmode: cfg.dns.blockingmode,
		blockingipv4: cfg.dns.blockingipv4,
		blockingipv6: cfg.dns.blockingipv6,
		rateLimitCount: String(cfg.dns.ratelimit.count),
		rateLimitInterval: String(cfg.dns.ratelimit.interval),
		revServerActive: cfg.dns.revServer.active,
		revServerCIDR: cfg.dns.revServer.cidr,
		revServerTarget: cfg.dns.revServer.target,
		revServerDomain: cfg.dns.revServer.domain,
		piholePTR: cfg.dns.piholePTR,
		querylogEnabled: cfg.dns.querylog.enabled,
		privacylevel: String(cfg.misc.privacylevel),
		queryDisplay: cfg.ftl.query_display,
		maxDBdays: String(cfg.ftl.database.maxDBdays),
		DBinterval: String(cfg.ftl.database.DBinterval),
		excludeClients: cfg.webserver.api.excludeClients.join('\n'),
		excludeDomains: cfg.webserver.api.excludeDomains.join('\n'),
		maxHistory: String(cfg.webserver.api.maxHistory),
		resolveIPv4: cfg.resolver.resolveIPv4,
		resolveIPv6: cfg.resolver.resolveIPv6,
		networkNames: cfg.resolver.networkNames,
	};
}

function formToPatch(form: ConfigFormState, original: PiholeConfig): PatchConfigRequest {
	const patch: PatchConfigRequest = {};
	const dns: PatchConfigRequest['dns'] = {};
	const misc: PatchConfigRequest['misc'] = {};
	const ftl: PatchConfigRequest['ftl'] = {};
	const webserver: PatchConfigRequest['webserver'] = { api: {} };
	const resolver: PatchConfigRequest['resolver'] = {};

	const newUpstreams = form.upstreams.split('\n').map((s) => s.trim()).filter(Boolean);
	if (JSON.stringify(newUpstreams) !== JSON.stringify(original.dns.upstreams)) dns.upstreams = newUpstreams;
	if (form.interface !== original.dns.interface) dns.interface = form.interface;
	const port = parseInt(form.port, 10);
	if (!isNaN(port) && port !== original.dns.port) dns.port = port;
	if (form.dnssec !== original.dns.dnssec) dns.dnssec = form.dnssec;
	if (form.domainNeeded !== original.dns.domainNeeded) dns.domainNeeded = form.domainNeeded;
	if (form.expandHosts !== original.dns.expandHosts) dns.expandHosts = form.expandHosts;
	const localTTL = parseInt(form.localTTL, 10);
	if (!isNaN(localTTL) && localTTL !== original.dns.localTTL) dns.localTTL = localTTL;
	if (form.blockingmode !== original.dns.blockingmode) dns.blockingmode = form.blockingmode;
	if (form.blockingipv4 !== original.dns.blockingipv4) dns.blockingipv4 = form.blockingipv4;
	if (form.blockingipv6 !== original.dns.blockingipv6) dns.blockingipv6 = form.blockingipv6;

	const rlCount = parseInt(form.rateLimitCount, 10);
	const rlInterval = parseInt(form.rateLimitInterval, 10);
	if (!isNaN(rlCount) && !isNaN(rlInterval) &&
		(rlCount !== original.dns.ratelimit.count || rlInterval !== original.dns.ratelimit.interval)) {
		dns.ratelimit = { count: rlCount, interval: rlInterval };
	}

	if (form.revServerActive !== original.dns.revServer.active ||
		form.revServerCIDR !== original.dns.revServer.cidr ||
		form.revServerTarget !== original.dns.revServer.target ||
		form.revServerDomain !== original.dns.revServer.domain) {
		dns.revServer = {
			active: form.revServerActive,
			cidr: form.revServerCIDR,
			target: form.revServerTarget,
			domain: form.revServerDomain,
		};
	}
	if (form.piholePTR !== original.dns.piholePTR) dns.piholePTR = form.piholePTR;
	if (form.querylogEnabled !== original.dns.querylog.enabled) dns.querylog = { enabled: form.querylogEnabled };

	if (Object.keys(dns).length > 0) patch.dns = dns;

	const pl = parseInt(form.privacylevel, 10);
	if (!isNaN(pl) && pl !== original.misc.privacylevel) misc.privacylevel = pl;
	if (Object.keys(misc).length > 0) patch.misc = misc;

	if (form.queryDisplay !== original.ftl.query_display) ftl.query_display = form.queryDisplay;
	const maxDBdays = parseInt(form.maxDBdays, 10);
	const dbInterval = parseFloat(form.DBinterval);
	if ((!isNaN(maxDBdays) && maxDBdays !== original.ftl.database.maxDBdays) ||
		(!isNaN(dbInterval) && dbInterval !== original.ftl.database.DBinterval)) {
		ftl.database = {};
		if (!isNaN(maxDBdays)) ftl.database.maxDBdays = maxDBdays;
		if (!isNaN(dbInterval)) ftl.database.DBinterval = dbInterval;
	}
	if (Object.keys(ftl).length > 0) patch.ftl = ftl;

	const newExcClients = form.excludeClients.split('\n').map((s) => s.trim()).filter(Boolean);
	const newExcDomains = form.excludeDomains.split('\n').map((s) => s.trim()).filter(Boolean);
	const maxHist = parseInt(form.maxHistory, 10);
	if (JSON.stringify(newExcClients) !== JSON.stringify(original.webserver.api.excludeClients))
		webserver.api!.excludeClients = newExcClients;
	if (JSON.stringify(newExcDomains) !== JSON.stringify(original.webserver.api.excludeDomains))
		webserver.api!.excludeDomains = newExcDomains;
	if (!isNaN(maxHist) && maxHist !== original.webserver.api.maxHistory)
		webserver.api!.maxHistory = maxHist;
	if (Object.keys(webserver.api!).length > 0) patch.webserver = webserver;

	if (form.resolveIPv4 !== original.resolver.resolveIPv4) resolver.resolveIPv4 = form.resolveIPv4;
	if (form.resolveIPv6 !== original.resolver.resolveIPv6) resolver.resolveIPv6 = form.resolveIPv6;
	if (form.networkNames !== original.resolver.networkNames) resolver.networkNames = form.networkNames;
	if (Object.keys(resolver).length > 0) patch.resolver = resolver;

	return patch;
}

// ---------- Sub-components ----------

function DriftBadge() {
	return <span className={styles.driftBadge} title='Value differs across nodes'>drift</span>;
}

function FieldRow({ label, drifted, children }: { label: string; drifted?: boolean; children: React.ReactNode }) {
	return (
		<div className={styles.fieldRow}>
			<div className={styles.fieldLabel}>
				{label}
				{drifted && <DriftBadge />}
			</div>
			<div className={styles.fieldControl}>{children}</div>
		</div>
	);
}

function SectionHeader({ title }: { title: string }) {
	return <h3 className={styles.sectionHeader}>{title}</h3>;
}

// ---------- Diff modal ----------

type DiffEntry = { label: string; current: string; proposed: string };

function buildDiff(patch: PatchConfigRequest, original: PiholeConfig): DiffEntry[] {
	const entries: DiffEntry[] = [];
	const fmt = (v: unknown) => (v === undefined || v === null) ? '—' : JSON.stringify(v);
	const add = (label: string, cur: unknown, next: unknown) => {
		if (JSON.stringify(cur) !== JSON.stringify(next)) entries.push({ label, current: fmt(cur), proposed: fmt(next) });
	};

	if (patch.dns) {
		const d = patch.dns;
		if (d.upstreams !== undefined) add('DNS Upstreams', original.dns.upstreams, d.upstreams);
		if (d.interface !== undefined) add('DNS Interface', original.dns.interface, d.interface);
		if (d.port !== undefined) add('DNS Port', original.dns.port, d.port);
		if (d.dnssec !== undefined) add('DNSSEC', original.dns.dnssec, d.dnssec);
		if (d.domainNeeded !== undefined) add('Domain Needed', original.dns.domainNeeded, d.domainNeeded);
		if (d.expandHosts !== undefined) add('Expand Hosts', original.dns.expandHosts, d.expandHosts);
		if (d.localTTL !== undefined) add('Local TTL', original.dns.localTTL, d.localTTL);
		if (d.blockingmode !== undefined) add('Blocking Mode', original.dns.blockingmode, d.blockingmode);
		if (d.blockingipv4 !== undefined) add('Blocking IPv4', original.dns.blockingipv4, d.blockingipv4);
		if (d.blockingipv6 !== undefined) add('Blocking IPv6', original.dns.blockingipv6, d.blockingipv6);
		if (d.ratelimit !== undefined) add('Rate Limit', original.dns.ratelimit, d.ratelimit);
		if (d.revServer !== undefined) add('Reverse Server', original.dns.revServer, d.revServer);
		if (d.piholePTR !== undefined) add('Pi-hole PTR', original.dns.piholePTR, d.piholePTR);
		if (d.querylog !== undefined) add('Query Log', original.dns.querylog, d.querylog);
	}
	if (patch.misc) {
		if (patch.misc.privacylevel !== undefined) add('Privacy Level', original.misc.privacylevel, patch.misc.privacylevel);
	}
	if (patch.ftl) {
		if (patch.ftl.query_display !== undefined) add('Query Display', original.ftl.query_display, patch.ftl.query_display);
		if (patch.ftl.database !== undefined) add('FTL Database', original.ftl.database, patch.ftl.database);
	}
	if (patch.webserver?.api) {
		const api = patch.webserver.api;
		if (api.excludeClients !== undefined) add('Exclude Clients', original.webserver.api.excludeClients, api.excludeClients);
		if (api.excludeDomains !== undefined) add('Exclude Domains', original.webserver.api.excludeDomains, api.excludeDomains);
		if (api.maxHistory !== undefined) add('Max History', original.webserver.api.maxHistory, api.maxHistory);
	}
	if (patch.resolver) {
		if (patch.resolver.resolveIPv4 !== undefined) add('Resolve IPv4', original.resolver.resolveIPv4, patch.resolver.resolveIPv4);
		if (patch.resolver.resolveIPv6 !== undefined) add('Resolve IPv6', original.resolver.resolveIPv6, patch.resolver.resolveIPv6);
		if (patch.resolver.networkNames !== undefined) add('Network Names', original.resolver.networkNames, patch.resolver.networkNames);
	}
	return entries;
}

// ---------- Config tab ----------

function PiholeConfigTab() {
	const { piholeNodes } = usePiholes();
	const [loading, setLoading] = useState(false);
	const [loadError, setLoadError] = useState<string | null>(null);
	const [original, setOriginal] = useState<PiholeConfig | null>(null);
	const [perNode, setPerNode] = useState<Record<string, PiholeConfig>>({});
	const [globalDrift, setGlobalDrift] = useState(false);
	const [form, setForm] = useState<ConfigFormState | null>(null);

	const [diffOpen, setDiffOpen] = useState(false);
	const [pendingPatch, setPendingPatch] = useState<PatchConfigRequest | null>(null);
	const [saving, setSaving] = useState(false);
	const [saveResult, setSaveResult] = useState<PatchConfigResponse | null>(null);
	const [saveError, setSaveError] = useState<string | null>(null);

	// Refs give the load callback stable access to current form/original without
	// recreating it on every keystroke (avoids useCallback deps churn).
	const formRef = useRef(form);
	const originalRef = useRef(original);
	useEffect(() => { formRef.current = form; originalRef.current = original; });

	const load = useCallback(async (force = false) => {
		if (!force && formRef.current && originalRef.current &&
			Object.keys(formToPatch(formRef.current, originalRef.current)).length > 0) {
			if (!window.confirm('You have unsaved changes. Reload and discard them?')) return;
		}
		setLoading(true);
		setLoadError(null);
		try {
			const resp = await getConfig();
			setGlobalDrift(resp.drifted);
			setPerNode(resp.perNode ?? {});
			if (resp.consensus) {
				setOriginal(resp.consensus);
				setForm(configToForm(resp.consensus));
			}
		} catch (err) {
			setLoadError(err instanceof Error ? err.message : 'Failed to load config');
			setSaveResult(null);
		} finally {
			setLoading(false);
		}
	}, []);

	useEffect(() => { load(true); }, [load]);

	function handleSave() {
		if (!form || !original) return;
		const patch = formToPatch(form, original);
		if (Object.keys(patch).length === 0) return;
		setPendingPatch(patch);
		setSaveResult(null);
		setSaveError(null);
		setDiffOpen(true);
	}

	async function handleConfirm() {
		if (!pendingPatch) return;
		setSaving(true);
		try {
			const result = await patchConfig(pendingPatch);
			setSaveResult(result);
			setDiffOpen(false);
			// Reload to reflect what Pi-hole actually stored
			await load(true);
		} catch (err) {
			setSaveError(err instanceof Error ? err.message : 'Failed to save config');
		} finally {
			setSaving(false);
		}
	}

	const set = (key: keyof ConfigFormState, val: string | boolean) =>
		setForm((f) => f ? { ...f, [key]: val } : f);

	const drift = (key: (c: PiholeConfig) => unknown) => isDrifted(perNode, key);

	const hasChanges = useMemo(
		() => !!(form && original && Object.keys(formToPatch(form, original)).length > 0),
		[form, original],
	);

	const diffEntries = useMemo(
		() => (pendingPatch && original ? buildDiff(pendingPatch, original) : []),
		[pendingPatch, original],
	);

	if (loading && !form) {
		return (
			<div className={styles.configLoading}>
				<RefreshCw size={20} className={styles.spin} />
				Loading config…
			</div>
		);
	}
	if (loadError && !form) {
		return <div className={styles.configError}>{loadError}</div>;
	}
	if (!form || !original) return null;

	return (
		<div className={styles.configTab}>
			{loadError && (
				<div className={styles.reloadErrorBanner}>
					<AlertTriangle size={14} />
					<span>Failed to reload: {loadError}</span>
					<button type='button' className={styles.reloadBtn} onClick={() => load(true)} disabled={loading}>Retry</button>
				</div>
			)}
			{globalDrift && (
				<div className={styles.driftBanner}>
					<AlertTriangle size={15} />
					Config values differ across nodes — drifted fields are highlighted below. Saving will overwrite all nodes.
				</div>
			)}

			{saveResult && (
				<div className={styles.saveResultBanner} data-partial={!Object.values(saveResult.nodes).every((n) => n.success) || undefined}>
					{Object.values(saveResult.nodes).every((n) => n.success) ? (
						<><CheckCircle size={14} /> Config saved to all nodes</>
					) : (
						<><AlertTriangle size={14} /> Partially saved — some nodes failed</>
					)}
					{Object.values(saveResult.nodes).map((n) => (
						<div key={n.node.id} className={styles.saveResultRow}>
							{n.success
								? <CheckCircle size={12} className={styles.iconOk} />
								: <XCircle size={12} className={styles.iconFail} />}
							<span>{n.node.name}</span>
							{n.error && <span className={styles.saveResultError}>{n.error}</span>}
						</div>
					))}
				</div>
			)}

			<div className={styles.configForm}>
				<SectionHeader title='DNS — Upstream servers' />
				<FieldRow label='Upstream resolvers' drifted={drift((c) => c.dns.upstreams)}>
					<textarea
						className={styles.textarea}
						rows={3}
						value={form.upstreams}
						onChange={(e) => set('upstreams', e.target.value)}
						placeholder='127.0.0.1#5335'
					/>
					<span className={styles.fieldHint}>One per line, e.g. 127.0.0.1#5335 or 8.8.8.8</span>
				</FieldRow>

				<SectionHeader title='DNS — Interface & port' />
				<FieldRow label='Interface' drifted={drift((c) => c.dns.interface)}>
					<input className={styles.input} value={form.interface} onChange={(e) => set('interface', e.target.value)} />
				</FieldRow>
				<FieldRow label='Port' drifted={drift((c) => c.dns.port)}>
					<input className={styles.inputShort} type='number' min={1} max={65535} value={form.port}
						onChange={(e) => set('port', e.target.value)} />
				</FieldRow>

				<SectionHeader title='DNS — Options' />
				<FieldRow label='DNSSEC' drifted={drift((c) => c.dns.dnssec)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.dnssec} onChange={(e) => set('dnssec', e.target.checked)} />
						<span>{form.dnssec ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Domain needed' drifted={drift((c) => c.dns.domainNeeded)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.domainNeeded} onChange={(e) => set('domainNeeded', e.target.checked)} />
						<span>{form.domainNeeded ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Expand hosts' drifted={drift((c) => c.dns.expandHosts)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.expandHosts} onChange={(e) => set('expandHosts', e.target.checked)} />
						<span>{form.expandHosts ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Local TTL (s)' drifted={drift((c) => c.dns.localTTL)}>
					<input className={styles.inputShort} type='number' min={0} value={form.localTTL}
						onChange={(e) => set('localTTL', e.target.value)} />
				</FieldRow>

				<SectionHeader title='DNS — Blocking' />
				<FieldRow label='Blocking mode' drifted={drift((c) => c.dns.blockingmode)}>
					<select className={styles.select} value={form.blockingmode} onChange={(e) => set('blockingmode', e.target.value)}>
						<option value='NULL'>NULL (recommended)</option>
						<option value='IP'>IP</option>
						<option value='IP-NODATA-AAAA'>IP-NODATA-AAAA</option>
						<option value='NXDOMAIN'>NXDOMAIN</option>
					</select>
				</FieldRow>
				{(form.blockingmode === 'IP' || form.blockingmode === 'IP-NODATA-AAAA') && (
					<>
						<FieldRow label='Blocking IPv4' drifted={drift((c) => c.dns.blockingipv4)}>
							<input className={styles.input} value={form.blockingipv4} onChange={(e) => set('blockingipv4', e.target.value)} />
						</FieldRow>
						<FieldRow label='Blocking IPv6' drifted={drift((c) => c.dns.blockingipv6)}>
							<input className={styles.input} value={form.blockingipv6} onChange={(e) => set('blockingipv6', e.target.value)} />
						</FieldRow>
					</>
				)}

				<SectionHeader title='DNS — Rate limiting' />
				<FieldRow label='Queries / interval' drifted={drift((c) => c.dns.ratelimit)}>
					<div className={styles.inlineRow}>
						<input className={styles.inputShort} type='number' min={0} value={form.rateLimitCount}
							onChange={(e) => set('rateLimitCount', e.target.value)} />
						<span className={styles.inlineSep}>queries per</span>
						<input className={styles.inputShort} type='number' min={1} value={form.rateLimitInterval}
							onChange={(e) => set('rateLimitInterval', e.target.value)} />
						<span className={styles.inlineSep}>seconds</span>
					</div>
				</FieldRow>

				<SectionHeader title='DNS — Reverse server' />
				<FieldRow label='Active' drifted={drift((c) => c.dns.revServer.active)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.revServerActive} onChange={(e) => set('revServerActive', e.target.checked)} />
						<span>{form.revServerActive ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				{form.revServerActive && (
					<>
						<FieldRow label='CIDR' drifted={drift((c) => c.dns.revServer.cidr)}>
							<input className={styles.input} value={form.revServerCIDR} onChange={(e) => set('revServerCIDR', e.target.value)} placeholder='192.168.0.0/24' />
						</FieldRow>
						<FieldRow label='Target' drifted={drift((c) => c.dns.revServer.target)}>
							<input className={styles.input} value={form.revServerTarget} onChange={(e) => set('revServerTarget', e.target.value)} placeholder='192.168.0.1' />
						</FieldRow>
						<FieldRow label='Domain' drifted={drift((c) => c.dns.revServer.domain)}>
							<input className={styles.input} value={form.revServerDomain} onChange={(e) => set('revServerDomain', e.target.value)} placeholder='local' />
						</FieldRow>
					</>
				)}

				<SectionHeader title='DNS — Pi-hole PTR' />
				<FieldRow label='Pi-hole PTR' drifted={drift((c) => c.dns.piholePTR)}>
					<select className={styles.select} value={form.piholePTR} onChange={(e) => set('piholePTR', e.target.value)}>
						<option value='PI.HOLE'>PI.HOLE — return "pi.hole"</option>
						<option value='HOSTNAME'>HOSTNAME — return machine hostname</option>
						<option value='HOSTNAMEFQDN'>HOSTNAMEFQDN — return FQDN</option>
						<option value='NONE'>NONE — disabled</option>
					</select>
				</FieldRow>

				<SectionHeader title='Query logging' />
				<FieldRow label='Query log' drifted={drift((c) => c.dns.querylog.enabled)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.querylogEnabled} onChange={(e) => set('querylogEnabled', e.target.checked)} />
						<span>{form.querylogEnabled ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Max DB retention (days)' drifted={drift((c) => c.ftl.database.maxDBdays)}>
					<input className={styles.inputShort} type='number' min={0} value={form.maxDBdays}
						onChange={(e) => set('maxDBdays', e.target.value)} />
				</FieldRow>
				<FieldRow label='DB write interval (min)' drifted={drift((c) => c.ftl.database.DBinterval)}>
					<input className={styles.inputShort} type='number' min={0} step='0.1' value={form.DBinterval}
						onChange={(e) => set('DBinterval', e.target.value)} />
				</FieldRow>

				<SectionHeader title='Privacy' />
				<FieldRow label='Privacy level' drifted={drift((c) => c.misc.privacylevel)}>
					<select className={styles.select} value={form.privacylevel}
						onChange={(e) => set('privacylevel', e.target.value)}>
						<option value='0'>0 — Show everything</option>
						<option value='1'>1 — Hide domains</option>
						<option value='2'>2 — Hide domains & clients</option>
						<option value='3'>3 — Anonymous mode</option>
					</select>
				</FieldRow>

				<SectionHeader title='API' />
				<FieldRow label='Exclude clients' drifted={drift((c) => c.webserver.api.excludeClients)}>
					<textarea className={styles.textarea} rows={3} value={form.excludeClients}
						onChange={(e) => set('excludeClients', e.target.value)}
						placeholder='One IP per line' />
				</FieldRow>
				<FieldRow label='Exclude domains' drifted={drift((c) => c.webserver.api.excludeDomains)}>
					<textarea className={styles.textarea} rows={3} value={form.excludeDomains}
						onChange={(e) => set('excludeDomains', e.target.value)}
						placeholder='One domain per line' />
				</FieldRow>
				<FieldRow label='Max history (h)' drifted={drift((c) => c.webserver.api.maxHistory)}>
					<input className={styles.inputShort} type='number' min={1} value={form.maxHistory}
						onChange={(e) => set('maxHistory', e.target.value)} />
				</FieldRow>

				<SectionHeader title='FTL' />
				<FieldRow label='Query display' drifted={drift((c) => c.ftl.query_display)}>
					<select className={styles.select} value={form.queryDisplay}
						onChange={(e) => set('queryDisplay', e.target.value)}>
						<option value='public'>public</option>
						<option value='hidden'>hidden</option>
						<option value='none'>none</option>
					</select>
				</FieldRow>

				<SectionHeader title='Resolver' />
				<FieldRow label='Resolve IPv4' drifted={drift((c) => c.resolver.resolveIPv4)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.resolveIPv4} onChange={(e) => set('resolveIPv4', e.target.checked)} />
						<span>{form.resolveIPv4 ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Resolve IPv6' drifted={drift((c) => c.resolver.resolveIPv6)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.resolveIPv6} onChange={(e) => set('resolveIPv6', e.target.checked)} />
						<span>{form.resolveIPv6 ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
				<FieldRow label='Network names' drifted={drift((c) => c.resolver.networkNames)}>
					<label className={styles.toggle}>
						<input type='checkbox' checked={form.networkNames} onChange={(e) => set('networkNames', e.target.checked)} />
						<span>{form.networkNames ? 'Enabled' : 'Disabled'}</span>
					</label>
				</FieldRow>
			</div>

			<div className={styles.configActions}>
				<button type='button' className={styles.reloadBtn} onClick={() => load(false)} disabled={loading}>
					<RefreshCw size={14} className={loading ? styles.spin : undefined} />
					Reload
				</button>
				<button type='button' className={styles.saveBtn} onClick={handleSave} disabled={!hasChanges || loading}>
					Save changes…
				</button>
			</div>

			{/* Diff / confirm modal */}
			{diffOpen && pendingPatch && (
				<div className={styles.modalOverlay}>
					<div className={styles.modal}>
						<h2 className={styles.modalTitle}>Review changes</h2>
						<p className={styles.modalSubtitle}>
							These changes will be applied to all {piholeNodes.length} node{piholeNodes.length !== 1 ? 's' : ''}.
						</p>
						{diffEntries.length === 0 ? (
							<p className={styles.modalEmpty}>No changes detected.</p>
						) : (
							<table className={styles.diffTable}>
								<thead>
									<tr>
										<th>Setting</th>
										<th>Current</th>
										<th>Proposed</th>
									</tr>
								</thead>
								<tbody>
									{diffEntries.map((e) => (
										<tr key={e.label}>
											<td className={styles.diffLabel}>{e.label}</td>
											<td className={styles.diffCurrent}>{e.current}</td>
											<td className={styles.diffProposed}>{e.proposed}</td>
										</tr>
									))}
								</tbody>
							</table>
						)}
						{saveError && <p className={styles.modalError}>{saveError}</p>}
						<div className={styles.modalActions}>
							<button type='button' className={styles.cancelBtn}
								onClick={() => { setDiffOpen(false); setPendingPatch(null); }}
								disabled={saving}>
								Cancel
							</button>
							<button type='button' className={styles.confirmBtn}
								onClick={handleConfirm}
								disabled={saving || diffEntries.length === 0}
								aria-busy={saving}>
								{saving ? <RefreshCw size={14} className={styles.spin} /> : null}
								{saving ? 'Applying…' : 'Apply to all nodes'}
							</button>
						</div>
					</div>
				</div>
			)}
		</div>
	);
}

// ---------- Main Settings ----------

export function Settings() {
	const [tab, setTab] = useState<Tab>('nodes');
	const nodesId = useId();
	const configId = useId();

	return (
		<div className={styles.settingsPage}>
			<div className={styles.tabBar} role='tablist'>
				<button
					id={`${nodesId}-tab`}
					role='tab'
					type='button'
					className={styles.tabBtn}
					data-active={tab === 'nodes' || undefined}
					onClick={() => setTab('nodes')}
					aria-selected={tab === 'nodes'}
					aria-controls={`${nodesId}-panel`}
				>
					Nodes
				</button>
				<button
					id={`${configId}-tab`}
					role='tab'
					type='button'
					className={styles.tabBtn}
					data-active={tab === 'pihole-config' || undefined}
					onClick={() => setTab('pihole-config')}
					aria-selected={tab === 'pihole-config'}
					aria-controls={`${configId}-panel`}
				>
					Pi-hole Config
				</button>
			</div>

			{tab === 'nodes' && (
				<div id={`${nodesId}-panel`} role='tabpanel' aria-labelledby={`${nodesId}-tab`}>
					<PiholeManagementList title='Pi-hole Nodes' />
				</div>
			)}
			{tab === 'pihole-config' && (
				<div id={`${configId}-panel`} role='tabpanel' aria-labelledby={`${configId}-tab`}>
					<PiholeConfigTab />
				</div>
			)}
		</div>
	);
}
