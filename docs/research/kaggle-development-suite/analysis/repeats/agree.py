import glob,json,os,sys,collections,statistics as st,random
def load(run):
    out={}
    for p in glob.glob(os.path.join(run,"*","*","0*.json")):
        j=json.load(open(p)); a=j.get("agent") or {}
        sc,cond,rep=p.split("/")[-3],p.split("/")[-2],os.path.basename(p)[:3]
        out[(sc,cond,rep)]=(float((j.get("grading") or {}).get("score") or 0),a.get("termination"),bool((j.get("grading") or {}).get("full_success")))
    return out
conds=["baseline","prompt","skill","rag"]
for name,rep_run,single_run in [(n,r,s) for n,r,s in zip(sys.argv[1::3],sys.argv[2::3],sys.argv[3::3])]:
    R=load(rep_run); S=load(single_run)
    scen=sorted({k[0] for k in R}); reps=sorted({k[2] for k in R})
    print(f"\n## {name}")
    print("cond      " + " ".join(f"rep{r[-1]:>5}" for r in reps) + "   mean   sd(rep)  single  in-range  timeouts")
    for c in conds:
        ms=[st.mean(R[(s,c,r)][0] for s in scen) for r in reps]
        sm=st.mean(S[(s,c,"001")][0] for s in scen)
        to=sum(R[(s,c,r)][1]=="timeout" for s in scen for r in reps)
        print(f"{c:9} "+" ".join(f"{m:8.3f}" for m in ms)+f"  {st.mean(ms):.3f}  {st.stdev(ms):.3f}   {sm:.3f}  {'yes' if min(ms)-1e-9<=sm<=max(ms)+1e-9 else 'no'}  {to}")
    # per-scenario consistency
    allsame=tot=0; agree=0; dif=[]
    for c in conds:
        for s in scen:
            f=[R[(s,c,r)][2] for r in reps]; tot+=1
            allsame+= len(set(f))==1
            maj=sum(f)>=2; agree+= (S[(s,c,"001")][2]==maj)
            dif.append(S[(s,c,"001")][0]-st.mean(R[(s,c,r)][0] for r in reps))
    print(f"scenario×condition cells: {tot}; all 3 repeats agree on full success: {allsame} ({allsame/tot:.0%}); single run matches repeat majority: {agree} ({agree/tot:.0%}); mean |single − repeat mean| per cell {st.mean(abs(x) for x in dif):.3f}")
    # condition ranking stability: prompt-vs-X per repeat
    for c in ["baseline","skill","rag"]:
        d=[st.mean(R[(s,c,r)][0]-R[(s,"prompt",r)][0] for s in scen) for r in reps]
        ds=st.mean(S[(s,c,"001")][0]-S[(s,"prompt","001")][0] for s in scen)
        print(f"  {c} − prompt per repeat: "+", ".join(f"{x:+.3f}" for x in d)+f"; single run {ds:+.3f}")
