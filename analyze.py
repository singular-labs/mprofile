#!/usr/bin/env python
import sys, time, os, os.path, sqlite3, json

prefix = os.path.dirname(sys.argv[0]) or '.'
f = open(prefix + '/vis.min.js')
js = f.read()
f.close()
f = open(prefix + '/vis.min.css')
css = f.read()
f.close()

f = open('mprofile.html', 'w')
print >> f, """<!doctype html>
<html>
<head>
<style>
%s
</style>
<script>
%s
</script>
</head>
<body>
<div id="timeline"></div>
<script>
var items = new vis.DataSet([
""" % (css, js)

db = sqlite3.connect('mprofile.sqlite')
db.row_factory = sqlite3.Row
c = db.cursor()
c.execute("""select * from profile where target<>'' order by started""")
prof = c.fetchall()

# headers = ['id', 'level', 'dir', 'pid', 'ppid', 'pppid', 'target', 'started',
#            'ended', 'status', 'cmd']

i = 0
g = 0
targets = dict() # ppid:target -> group
parents = dict() # p4id -> ppid:target

origin = prof[0]['started']
for row in prof[:-1]:
  i += 1
  item = {
    'id': i,
    'type': 'range',
    'title': row['cmd'],
    'content': str(row['ppid']) + ':' + row['target'],
    'start': (row['started'] - origin) * 1e-6,
    'end': row['ended'] * 1e-6
  }
  # if row['p3id'] in targets:
  #   print 'pid', row['pid'], 'derived3 from', row['p3id']
  # if row['p4id'] in targets:
  #   print 'pid', row['pid'], 'derived4 from', row['p4id']
  # if row['p5id'] in targets:
  #   print 'pid', row['pid'], 'derived5 from', row['p5id']
  groupkey = '%s:%s' % (row['ppid'], row['target'])
  parent_group = row['p4id']
  parents[row['pid']] = groupkey
  if groupkey in targets:
    item['group'] = targets[groupkey]['id']
  else:
    g += 1
    targets[groupkey] = {
      'id': g,
      #'content': str(row['ppid']) + ':' + row['target'],
      'content': row['target'],
      'nestedGroups': [],
      'showNested': False,
      'subgroupOrder': 'start'
    }
    item['group'] = g
    if parent_group in parents:
      pg = parents[parent_group]
      targets[pg]['nestedGroups'].append(g)
    
  print >> f, json.dumps(item),
  if i < len(prof) - 1:
    print >> f, ','

print >> f, """
]);
var groups = new vis.DataSet();
groups.add([
"""
i = 0
for g in sorted([v for v in targets.values() if type(v) == dict],
                key=lambda x:x['id']):
  i += 1
  if not g['nestedGroups']:
    del g['nestedGroups']
  print >> f, json.dumps(g),
  if i < len(targets):
    print >> f, ','
print >> f, """
]);
var options = {
  groupOrder: "start"
};
var timeline = new vis.Timeline(
  document.getElementById("timeline"),
  items,
  groups,
  options
);
</script>
</body>
</html>
"""
f.close()

if sys.platform == 'darwin':
  os.system('open mprofile.html')
