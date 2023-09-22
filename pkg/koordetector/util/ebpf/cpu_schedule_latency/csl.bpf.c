#include "vmlinux.h"
#include "bpf_helpers.h"
#include "bpf_core_read.h"

#define TASK_RUNNING 0
#define MAX_CGROUP_NAME_SIZE 128

struct {
	__uint(type, BPF_MAP_TYPE_LRU_HASH);
	__uint(max_entries, 65536);
	__type(key, u32);
	__type(value, u64);
} pid_start_time SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 1024);
	__type(key, char[MAX_CGROUP_NAME_SIZE]);
	__type(value, u64);
} output_cgroup_delay SEC(".maps");

struct {
	__uint(type, BPF_MAP_TYPE_HASH);
	__uint(max_entries, 1024);
	__type(key, char[MAX_CGROUP_NAME_SIZE]);
	__type(value, u64);
} output_cgroup_counter SEC(".maps");

struct sched_wakeup_tp_args {
	struct trace_entry ent;
	char comm[16];
	pid_t pid;
	int prio;
	int success;
	int target_cpu;
	char __data[0];
};

struct sched_switch_tp_args {
	struct trace_entry ent;
	char prev_comm[16];
	pid_t prev_pid;
	int prev_prio;
	long int prev_state;
	char next_comm[16];
	pid_t next_pid;
	int next_prio;
	char __data[0];
};

/* record enqueue timestamp */
static __always_inline
int trace_enqueue(struct task_struct *task)
{
    u64 ts;
	u32 pid;

	ts = bpf_ktime_get_ns();
	// task->tgid is thread group id which is userspace pid
	bpf_core_read(&pid, sizeof(pid), &task->tgid);

	bpf_map_update_elem(&pid_start_time, &pid, &ts, 0);
	return 0;
}

SEC("tp/sched/sched_wakeup")
int handle__sched_wakeup(struct sched_wakeup_tp_args *ctx)
{
    struct task_struct *task = (void *)bpf_get_current_task();

	return trace_enqueue(task);
}

SEC("tp/sched/sched_wakeup_new")
int handle__sched_wakeup_new(struct sched_wakeup_tp_args *ctx)
{
    struct task_struct *task = (void *)bpf_get_current_task();

    return trace_enqueue(task);
}

SEC("tp/sched/sched_switch")
int handle_switch(struct sched_switch_tp_args *ctx)
{
    struct task_struct *task = (void *)bpf_get_current_task();
    struct css_set *cgroups;
    struct cgroup_subsys_state *subsys[14];
    struct cgroup *cg;
    struct kernfs_node *kn;
    char *cgroup_name;
	char cgroup_name_array[MAX_CGROUP_NAME_SIZE];
	long name_len;

    bpf_core_read(&cgroups, sizeof(cgroups), &task->cgroups);
    bpf_core_read(&subsys, sizeof(subsys), &cgroups->subsys);
    bpf_core_read(&cg, sizeof(cg), &subsys[1]->cgroup);
    bpf_core_read(&kn, sizeof(kn), &cg->kn);
    bpf_core_read(&cgroup_name,sizeof(cgroup_name),&kn->name);
    if (!cgroup_name)
        return 0;
    name_len = bpf_core_read_str(&cgroup_name_array, MAX_CGROUP_NAME_SIZE, cgroup_name);
    if (name_len < 0)
        return 0;

	u32 pid = ctx->next_pid;
	long int prev_state;
	/* ivcsw: treat like an enqueue event and store timestamp */
	prev_state = ctx->prev_state;
	if (prev_state == TASK_RUNNING) {
	    struct task_struct *prev = (void *)bpf_get_current_task();
    	trace_enqueue(prev);
	}

    u64 *tsp, delta, now;
    u64 *last, *lastcounter;
    u64 init_delay = 0;
    u64 init_counter = 0;
	/* fetch timestamp and calculate delta */
	tsp = bpf_map_lookup_elem(&pid_start_time, &pid);
	if (!tsp)
		return 0;   /* missed enqueue */
	now = bpf_ktime_get_ns();
	delta = (now - *tsp);

    /* sum deltas and count switch times in this collect period*/
    lastcounter = bpf_map_lookup_elem(&output_cgroup_counter, &cgroup_name_array);
    if (!lastcounter)
        lastcounter = &init_counter;
    last = bpf_map_lookup_elem(&output_cgroup_delay, &cgroup_name_array);
    if (!last)
        last = &init_delay;
    *lastcounter = *lastcounter + 1;
    delta = (delta + *last);

    /* update maps*/
    bpf_map_update_elem(&output_cgroup_counter, &cgroup_name_array, lastcounter, BPF_ANY);
    bpf_map_update_elem(&output_cgroup_delay, &cgroup_name_array, &delta, BPF_ANY);
	bpf_map_delete_elem(&pid_start_time, &pid);

	return 0;
}

char LICENSE[] SEC("license") = "GPL";
