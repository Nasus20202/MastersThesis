import glob,json,sys,os
for arg in sys.argv[1:]:
    label,d=arg.split('=',1); dn=da=pn=pc=cached=prompt=0
    for p in glob.glob(os.path.join(d,'*','*','*.json')):
        a=json.load(open(p)).get('agent') or {}
        for r in a.get('responses') or []:
            t=(r.get('response') or {}).get('timings') or {}; u=(r.get('response') or {}).get('usage') or {}
            dn+=t.get('draft_n') or 0; da+=t.get('draft_n_accepted') or 0
            pn+=t.get('prompt_n') or 0; prompt+=u.get('prompt_tokens') or 0; cached+=u.get('cached_tokens') or 0
    print(f"{label:10} draft acceptance {da/max(dn,1):.3f}  prompt tokens evaluated {pn/max(prompt,1):.3f} of total (cache hit {cached/max(prompt,1):.3f})")
