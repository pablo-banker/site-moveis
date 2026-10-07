#!/usr/bin/env python3
"""Exercise production config with disposable TLS/SMTP/DB fixtures; no public deployment."""
import base64, http.client, json, os, re, secrets, socket, ssl, subprocess, tempfile, time
from pathlib import Path
ROOT=Path(__file__).resolve().parents[1]
def run(args, **kw):
 result=subprocess.run(args,cwd=ROOT,text=True,capture_output=True,**kw)
 if result.returncode: raise RuntimeError(f"Command failed: {args[0:3]}\n{result.stderr[-1500:]}")
 return result.stdout
with tempfile.TemporaryDirectory(prefix='forma-release-') as temporary:
 directory=Path(temporary);directory.chmod(0o755)
 password=secrets.token_hex(24);secret=secrets.token_hex(32);metrics=secrets.token_hex(32)
 environment={'SITE_HOST':'forma.localhost','TLS_EMAIL':'test@example.invalid','DATABASE_URL':f'postgres://forma_runtime:{password}@tlsdb/forma?sslmode=verify-full&sslrootcert=/certs/ca.crt','MIGRATION_DATABASE_URL':f'postgres://migrator:{password}@tlsdb/forma?sslmode=verify-full&sslrootcert=/certs/ca.crt','RUNTIME_DB_ROLE':'forma_runtime','BFF_SHARED_SECRET':secret,'METRICS_TOKEN':metrics,'SMTP_HOST':'mailpit','SMTP_PORT':'1025','SMTP_TLS_MODE':'starttls','SMTP_FROM':'Forma <conta@example.invalid>','EMAIL_OUTBOX_KEY':base64.b64encode(secrets.token_bytes(32)).decode(),'BACKUP_DIRECTORY_HOST':str(directory/'backups')}
 envfile=directory/'test.env';envfile.write_text('\n'.join(f'{k}={v}' for k,v in environment.items()));envfile.chmod(0o600)
 config=json.loads(run(['docker','compose','--env-file',str(envfile),'-f','compose.production.yaml','--profile','*','config','--format','json']))
 config['name']='forma-release-test'
 for group in ['networks','volumes']:
  for settings in config.get(group,{}).values():settings.pop('name',None)
 certs=directory/'certs';certs.mkdir()
 run(['openssl','req','-x509','-newkey','rsa:2048','-nodes','-keyout',str(certs/'ca.key'),'-out',str(certs/'ca.crt'),'-days','1','-subj','/CN=Forma release test CA'])
 run(['openssl','req','-newkey','rsa:2048','-nodes','-keyout',str(certs/'server.key'),'-out',str(certs/'server.csr'),'-subj','/CN=tlsdb'])
 (certs/'extensions').write_text('subjectAltName=DNS:tlsdb,DNS:mailpit\nextendedKeyUsage=serverAuth\n')
 run(['openssl','x509','-req','-in',str(certs/'server.csr'),'-CA',str(certs/'ca.crt'),'-CAkey',str(certs/'ca.key'),'-CAcreateserial','-out',str(certs/'server.crt'),'-days','1','-extfile',str(certs/'extensions')])
 # Fixture keys are disposable, accessible only in this isolated test network.
 certs.chmod(0o755)
 for p in certs.iterdir():p.chmod(0o644)
 init=directory/'init.sql';init.write_text(f"CREATE ROLE forma_runtime LOGIN PASSWORD '{password}';\n")
 caddy=directory/'Caddyfile';caddy.write_text('https://forma.localhost {\n tls internal\n reverse_proxy web:3000 {\n header_up X-Forwarded-For {remote_host}\n }\n}\n')
 bind=lambda source,target:{'type':'bind','source':str(source),'target':target,'read_only':True}
 for service in ['api','migrate']:
  config['services'][service].pop('build',None);config['services'][service]['image']='forma-api'
  config['services'][service].setdefault('volumes',[]).append(bind(certs,'/certs'))
  config['services'][service]['environment']['SSL_CERT_FILE']='/certs/ca.crt'
 config['services']['web'].pop('build',None);config['services']['web']['image']='forma-web:release-check'
 config['services']['backup'].pop('build',None);config['services']['backup']['image']='forma-backup:release-check';config['services']['backup'].setdefault('volumes',[]).append(bind(certs,'/certs'))
 config['services']['backup']['environment']['BACKUP_DATABASE_URL']=environment['MIGRATION_DATABASE_URL']
 config['services']['caddy']['volumes']=[v for v in config['services']['caddy']['volumes'] if v['target']!='/etc/caddy/Caddyfile']+[bind(caddy,'/etc/caddy/Caddyfile')]
 config['services']['caddy']['ports']=[{'target':443,'published':'0','host_ip':'127.0.0.1','protocol':'tcp'}]
 config['services']['tlsdb']={'image':'postgres:17-alpine@sha256:b0f9560a2de083e2cc7382e75f808c7381a32852a7ec49117deedb300e552b24','environment':{'POSTGRES_USER':'migrator','POSTGRES_PASSWORD':password,'POSTGRES_DB':'forma'},'networks':['outbound'],'volumes':[bind(certs,'/certs'),bind(init,'/docker-entrypoint-initdb.d/runtime.sql')],'entrypoint':['sh','-c','cp /certs/server.key /tmp/server.key; chown postgres /tmp/server.key; chmod 600 /tmp/server.key; exec docker-entrypoint.sh postgres -c ssl=on -c ssl_cert_file=/certs/server.crt -c ssl_key_file=/tmp/server.key'],'healthcheck':{'test':['CMD-SHELL','pg_isready -U migrator -d forma'],'interval':'2s','timeout':'2s','retries':30}}
 config['services']['mailpit']={'image':'axllent/mailpit@sha256:b68349e3a014b90c5610bfb26b2ae36f3892d7b8cf25ee140c6c71c98d2fcf48','environment':{'MP_SMTP_TLS_CERT':'/certs/server.crt','MP_SMTP_TLS_KEY':'/certs/server.key','MP_SMTP_REQUIRE_STARTTLS':'true'},'volumes':[bind(certs,'/certs')],'networks':['outbound'],'ports':[{'target':8025,'published':'0','host_ip':'127.0.0.1','protocol':'tcp'}]}
 file=directory/'compose.json';file.write_text(json.dumps(config));file.chmod(0o600)
 command=['docker','compose','-p','forma-release-test','-f',str(file)]
 def compose(*args):return run(command+list(args))
 try:
  compose('up','-d','--wait','tlsdb','mailpit')
  compose('run','--rm','migrate')
  compose('up','-d','--wait','--wait-timeout','90','api','web','caddy')
  root=directory/'caddy-root.crt';compose('cp','caddy:/data/caddy/pki/authorities/local/root.crt',str(root))
  port=int(compose('port','caddy','443').strip().split(':')[-1]);mailport=int(compose('port','mailpit','8025').strip().split(':')[-1])
  trust=ssl.create_default_context(cafile=str(root));cookie=''
  def request(method,path,body=None,expected=200,origin='https://forma.localhost'):
   conn=http.client.HTTPSConnection('forma.localhost',context=trust,timeout=10)
   conn.sock=trust.wrap_socket(socket.create_connection(('127.0.0.1',port),timeout=10),server_hostname='forma.localhost')
   headers={'Host':'forma.localhost','Origin':origin,'Content-Type':'application/json','Cookie':cookie}
   conn.request(method,path,json.dumps(body) if body is not None else None,headers)
   response=conn.getresponse();raw=response.read();assert response.status==expected,(path,response.status,raw[:300]);head=dict(response.getheaders());conn.close();return (json.loads(raw) if raw.startswith(b'{') else raw,head)
  home,headers=request('GET','/');assert 'max-age=' in headers.get('Strict-Transport-Security','');assert "script-src 'self'" in headers.get('Content-Security-Policy','')
  request('POST','/backend/auth/logout',{},403,origin='https://attacker.invalid')
  user={'name':'Release test','email':'release@example.invalid','password':'ReleaseTestPassword123!'}
  request('POST','/backend/auth/register',user,202)
  def delivered(subject):
   for _ in range(45):
    conn=http.client.HTTPConnection('127.0.0.1',mailport);conn.request('GET','/api/v1/messages');r=conn.getresponse();messages=json.loads(r.read());conn.close()
    for item in messages.get('messages',[]):
     if item['Subject']==subject:
      conn=http.client.HTTPConnection('127.0.0.1',mailport);conn.request('GET','/api/v1/message/'+item['ID']);r=conn.getresponse();message=json.loads(r.read());conn.close();return re.search(r'#token=([A-Za-z0-9_-]{43})',message['Text']).group(1)
    time.sleep(1)
   raise AssertionError('SMTP delivery timeout')
  token=delivered('Confirme seu e-mail na Forma');request('POST','/backend/auth/verify/confirm',{'token':token})
  _,headers=request('POST','/backend/auth/login',{'email':user['email'],'password':user['password']});setcookie=headers.get('Set-Cookie','');assert '__Host-forma-session=' in setcookie and 'Secure' in setcookie and 'HttpOnly' in setcookie;cookie=setcookie.split(';')[0]
  current,_=request('GET','/backend/auth/me');assert current['emailVerified']
  items=[{'id':'sofa-arco','finish':'Linho natural','quantity':1}]
  quote,_=request('POST','/backend/shipping/quotes',{'cep':'01001000','items':items})
  import uuid
  order,_=request('POST','/backend/orders',{'idempotencyKey':str(uuid.uuid4()),'shippingQuoteId':quote['id'],'items':items,'address':{'cep':'01001000','street':'Rua Teste','number':'10','city':'São Paulo','state':'SP'}},201)
  paid,_=request('POST',f"/backend/orders/{order['id']}/payment",{});assert paid['status']=='paid'
  request('POST','/backend/auth/reset/request',{'email':user['email']});reset=delivered('Redefina sua senha na Forma')
  request('POST','/backend/auth/reset/confirm',{'token':reset,'password':'UpdatedReleasePassword123!'});request('GET','/backend/auth/me',expected=401)
  # Metrics are private and authenticated. SQL verifies low privilege at API startup.
  result=compose('exec','-T','api','wget','-q','-O','-','--header=Authorization: Bearer '+metrics,'http://127.0.0.1:8081/internal/metrics');assert 'forma_email_pending 0' in result
  backupdir=directory/'backups';backupdir.mkdir(exist_ok=True)
  recipient=compose('run','--rm','--entrypoint','sh','backup','-c','age-keygen -o /backups/test.agekey >/dev/null 2>&1; age-keygen -y /backups/test.agekey').strip()
  compose('run','--rm','-e','BACKUP_AGE_RECIPIENT='+recipient,'backup')
  archives=list(backupdir.glob('*.dump.age'));assert len(archives)==1
  compose('exec','-T','tlsdb','psql','-U','migrator','-d','postgres','-c','CREATE DATABASE forma_restore;')
  restore_url=environment['MIGRATION_DATABASE_URL'].replace('/forma?','/forma_restore?')
  compose('run','--rm','-e','RESTORE_DATABASE_URL='+restore_url,'-e','RESTORE_AGE_IDENTITY=/backups/test.agekey','-e','CONFIRM_RESTORE=isolated-target','--entrypoint','/usr/local/bin/restore.sh','backup','/backups/'+archives[0].name)
  counts=compose('exec','-T','tlsdb','psql','-U','migrator','-d','forma_restore','-Atc','SELECT count(*) FROM customers;SELECT count(*) FROM orders;').strip();assert counts=='1\n1',counts
  print('Production smoke passed: HTTPS/HSTS, verified Postgres TLS, restricted runtime role, BFF signing, Secure cookies, SMTP STARTTLS, verification, checkout, simulated payment, password reset, metrics and encrypted backup/restore.')
 finally:
  compose('down','--volumes','--remove-orphans')
