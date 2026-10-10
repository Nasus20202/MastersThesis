import csv,statistics as st,collections,sys
rows=list(csv.DictReader(open(sys.argv[1])))
g=collections.defaultdict(list)
for r in rows: g[r['label']].append(r)
print(f"{'label':14} {'n':>4} {'macro':>6} {'full':>4} {'tmo':>4} {'turnl':>5} {'err':>4} {'dur':>6} {'queue':>6} {'prefl':>6} {'decode':>6} {'tok/s':>6} {'turns':>5} {'peak':>6}")
for k,v in g.items():
    m=lambda key: st.mean(float(r[key]) for r in v if r[key] not in ('',))
    t=collections.Counter(r['termination'] for r in v)
    full=sum(r['full']=='True' for r in v)
    print(f"{k:14} {len(v):4} {m('score'):6.3f} {full:4} {t['timeout']:4} {t['turn_limit']:5} {t['inference_error']+t['error']:4} {m('duration_s'):6.0f} {m('queue_s'):6.0f} {m('prefill_s'):6.0f} {m('decode_s'):6.0f} {m('decode_tps'):6.1f} {m('turns'):5.1f} {m('peak_context'):6.0f}")
