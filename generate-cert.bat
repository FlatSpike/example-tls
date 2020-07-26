set PATH_OUT=.\certificates

if not exist %PATH_OUT% mkdir %PATH_OUT%

:: Generate Certificate Authority
:: Generate Private Key
openssl genrsa -out %PATH_OUT%\ca.key 2048

:: Generate Certificate
openssl req -new -x509^
  -key %PATH_OUT%\ca.key^
  -out %PATH_OUT%\ca.crt^
  -config .\ca.ini

:: Generate Certificate
:: Generate Private Key
openssl genrsa -out %PATH_OUT%\cert.key 2048

:: Generate CSR
openssl req -new^
  -key %PATH_OUT%\cert.key^
  -out %PATH_OUT%\cert.csr^
  -config .\cert.ini

:: Sign Cert
openssl x509 -req^
  -in %PATH_OUT%\cert.csr^
  -CA %PATH_OUT%\ca.crt^
  -CAkey %PATH_OUT%\ca.key^
  -CAcreateserial -out %PATH_OUT%\cert.crt^
  -extensions v3_req^
  -extfile .\cert.ini