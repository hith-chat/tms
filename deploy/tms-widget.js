(function(A,O){typeof exports=="object"&&typeof module<"u"?O(exports):typeof define=="function"&&define.amd?define(["exports"],O):(A=typeof globalThis<"u"?globalThis:A||self,O(A.TMSChatWidget={}))})(this,function(A){"use strict";class O{constructor(){this.events={}}on(e,t){this.events[e]||(this.events[e]=[]),this.events[e].push(t)}off(e,t){const i=this.events[e];if(i){const o=i.indexOf(t);o>-1&&i.splice(o,1)}}emit(e,t){const i=this.events[e];i&&i.forEach(o=>o(t))}}const H=crypto,G=s=>s instanceof CryptoKey,w=new TextEncoder,I=new TextDecoder;function j(...s){const e=s.reduce((o,{length:n})=>o+n,0),t=new Uint8Array(e);let i=0;for(const o of s)t.set(o,i),i+=o.length;return t}const Te=s=>{let e=s;typeof e=="string"&&(e=w.encode(e));const t=32768,i=[];for(let o=0;o<e.length;o+=t)i.push(String.fromCharCode.apply(null,e.subarray(o,o+t)));return btoa(i.join(""))},N=s=>Te(s).replace(/=/g,"").replace(/\+/g,"-").replace(/\//g,"_"),Ae=s=>{const e=atob(s),t=new Uint8Array(e.length);for(let i=0;i<e.length;i++)t[i]=e.charCodeAt(i);return t},E=s=>{let e=s;e instanceof Uint8Array&&(e=I.decode(e)),e=e.replace(/-/g,"+").replace(/_/g,"/").replace(/\s/g,"");try{return Ae(e)}catch{throw new TypeError("The input to be decoded is not correctly encoded.")}};class y extends Error{constructor(e,t){var i;super(e,t),this.code="ERR_JOSE_GENERIC",this.name=this.constructor.name,(i=Error.captureStackTrace)==null||i.call(Error,this,this.constructor)}}y.code="ERR_JOSE_GENERIC";class b extends y{constructor(e,t,i="unspecified",o="unspecified"){super(e,{cause:{claim:i,reason:o,payload:t}}),this.code="ERR_JWT_CLAIM_VALIDATION_FAILED",this.claim=i,this.reason=o,this.payload=t}}b.code="ERR_JWT_CLAIM_VALIDATION_FAILED";class B extends y{constructor(e,t,i="unspecified",o="unspecified"){super(e,{cause:{claim:i,reason:o,payload:t}}),this.code="ERR_JWT_EXPIRED",this.claim=i,this.reason=o,this.payload=t}}B.code="ERR_JWT_EXPIRED";class X extends y{constructor(){super(...arguments),this.code="ERR_JOSE_ALG_NOT_ALLOWED"}}X.code="ERR_JOSE_ALG_NOT_ALLOWED";class S extends y{constructor(){super(...arguments),this.code="ERR_JOSE_NOT_SUPPORTED"}}S.code="ERR_JOSE_NOT_SUPPORTED";class Ie extends y{constructor(e="decryption operation failed",t){super(e,t),this.code="ERR_JWE_DECRYPTION_FAILED"}}Ie.code="ERR_JWE_DECRYPTION_FAILED";class Ce extends y{constructor(){super(...arguments),this.code="ERR_JWE_INVALID"}}Ce.code="ERR_JWE_INVALID";class u extends y{constructor(){super(...arguments),this.code="ERR_JWS_INVALID"}}u.code="ERR_JWS_INVALID";class K extends y{constructor(){super(...arguments),this.code="ERR_JWT_INVALID"}}K.code="ERR_JWT_INVALID";class ke extends y{constructor(){super(...arguments),this.code="ERR_JWK_INVALID"}}ke.code="ERR_JWK_INVALID";class We extends y{constructor(){super(...arguments),this.code="ERR_JWKS_INVALID"}}We.code="ERR_JWKS_INVALID";class Re extends y{constructor(e="no applicable key found in the JSON Web Key Set",t){super(e,t),this.code="ERR_JWKS_NO_MATCHING_KEY"}}Re.code="ERR_JWKS_NO_MATCHING_KEY";class Me extends y{constructor(e="multiple matching keys found in the JSON Web Key Set",t){super(e,t),this.code="ERR_JWKS_MULTIPLE_MATCHING_KEYS"}}Me.code="ERR_JWKS_MULTIPLE_MATCHING_KEYS";class Oe extends y{constructor(e="request timed out",t){super(e,t),this.code="ERR_JWKS_TIMEOUT"}}Oe.code="ERR_JWKS_TIMEOUT";class Z extends y{constructor(e="signature verification failed",t){super(e,t),this.code="ERR_JWS_SIGNATURE_VERIFICATION_FAILED"}}Z.code="ERR_JWS_SIGNATURE_VERIFICATION_FAILED";function v(s,e="algorithm.name"){return new TypeError(`CryptoKey does not support this operation, its ${e} must be ${s}`)}function $(s,e){return s.name===e}function L(s){return parseInt(s.name.slice(4),10)}function $e(s){switch(s){case"ES256":return"P-256";case"ES384":return"P-384";case"ES512":return"P-521";default:throw new Error("unreachable")}}function De(s,e){if(e.length&&!e.some(t=>s.usages.includes(t))){let t="CryptoKey does not support this operation, its usages must include ";if(e.length>2){const i=e.pop();t+=`one of ${e.join(", ")}, or ${i}.`}else e.length===2?t+=`one of ${e[0]} or ${e[1]}.`:t+=`${e[0]}.`;throw new TypeError(t)}}function He(s,e,...t){switch(e){case"HS256":case"HS384":case"HS512":{if(!$(s.algorithm,"HMAC"))throw v("HMAC");const i=parseInt(e.slice(2),10);if(L(s.algorithm.hash)!==i)throw v(`SHA-${i}`,"algorithm.hash");break}case"RS256":case"RS384":case"RS512":{if(!$(s.algorithm,"RSASSA-PKCS1-v1_5"))throw v("RSASSA-PKCS1-v1_5");const i=parseInt(e.slice(2),10);if(L(s.algorithm.hash)!==i)throw v(`SHA-${i}`,"algorithm.hash");break}case"PS256":case"PS384":case"PS512":{if(!$(s.algorithm,"RSA-PSS"))throw v("RSA-PSS");const i=parseInt(e.slice(2),10);if(L(s.algorithm.hash)!==i)throw v(`SHA-${i}`,"algorithm.hash");break}case"EdDSA":{if(s.algorithm.name!=="Ed25519"&&s.algorithm.name!=="Ed448")throw v("Ed25519 or Ed448");break}case"Ed25519":{if(!$(s.algorithm,"Ed25519"))throw v("Ed25519");break}case"ES256":case"ES384":case"ES512":{if(!$(s.algorithm,"ECDSA"))throw v("ECDSA");const i=$e(e);if(s.algorithm.namedCurve!==i)throw v(i,"algorithm.namedCurve");break}default:throw new TypeError("CryptoKey does not support this operation")}De(s,t)}function Q(s,e,...t){var i;if(t=t.filter(Boolean),t.length>2){const o=t.pop();s+=`one of type ${t.join(", ")}, or ${o}.`}else t.length===2?s+=`one of type ${t[0]} or ${t[1]}.`:s+=`of type ${t[0]}.`;return e==null?s+=` Received ${e}`:typeof e=="function"&&e.name?s+=` Received function ${e.name}`:typeof e=="object"&&e!=null&&(i=e.constructor)!=null&&i.name&&(s+=` Received an instance of ${e.constructor.name}`),s}const ee=(s,...e)=>Q("Key must be ",s,...e);function te(s,e,...t){return Q(`Key for the ${s} algorithm must be `,e,...t)}const se=s=>G(s)?!0:(s==null?void 0:s[Symbol.toStringTag])==="KeyObject",J=["CryptoKey"],ie=(...s)=>{const e=s.filter(Boolean);if(e.length===0||e.length===1)return!0;let t;for(const i of e){const o=Object.keys(i);if(!t||t.size===0){t=new Set(o);continue}for(const n of o){if(t.has(n))return!1;t.add(n)}}return!0};function Ke(s){return typeof s=="object"&&s!==null}function C(s){if(!Ke(s)||Object.prototype.toString.call(s)!=="[object Object]")return!1;if(Object.getPrototypeOf(s)===null)return!0;let e=s;for(;Object.getPrototypeOf(e)!==null;)e=Object.getPrototypeOf(e);return Object.getPrototypeOf(s)===e}const oe=(s,e)=>{if(s.startsWith("RS")||s.startsWith("PS")){const{modulusLength:t}=e.algorithm;if(typeof t!="number"||t<2048)throw new TypeError(`${s} requires key modulusLength to be 2048 bits or larger`)}};function k(s){return C(s)&&typeof s.kty=="string"}function Je(s){return s.kty!=="oct"&&typeof s.d=="string"}function Pe(s){return s.kty!=="oct"&&typeof s.d>"u"}function Ne(s){return k(s)&&s.kty==="oct"&&typeof s.k=="string"}function Be(s){let e,t;switch(s.kty){case"RSA":{switch(s.alg){case"PS256":case"PS384":case"PS512":e={name:"RSA-PSS",hash:`SHA-${s.alg.slice(-3)}`},t=s.d?["sign"]:["verify"];break;case"RS256":case"RS384":case"RS512":e={name:"RSASSA-PKCS1-v1_5",hash:`SHA-${s.alg.slice(-3)}`},t=s.d?["sign"]:["verify"];break;case"RSA-OAEP":case"RSA-OAEP-256":case"RSA-OAEP-384":case"RSA-OAEP-512":e={name:"RSA-OAEP",hash:`SHA-${parseInt(s.alg.slice(-3),10)||1}`},t=s.d?["decrypt","unwrapKey"]:["encrypt","wrapKey"];break;default:throw new S('Invalid or unsupported JWK "alg" (Algorithm) Parameter value')}break}case"EC":{switch(s.alg){case"ES256":e={name:"ECDSA",namedCurve:"P-256"},t=s.d?["sign"]:["verify"];break;case"ES384":e={name:"ECDSA",namedCurve:"P-384"},t=s.d?["sign"]:["verify"];break;case"ES512":e={name:"ECDSA",namedCurve:"P-521"},t=s.d?["sign"]:["verify"];break;case"ECDH-ES":case"ECDH-ES+A128KW":case"ECDH-ES+A192KW":case"ECDH-ES+A256KW":e={name:"ECDH",namedCurve:s.crv},t=s.d?["deriveBits"]:[];break;default:throw new S('Invalid or unsupported JWK "alg" (Algorithm) Parameter value')}break}case"OKP":{switch(s.alg){case"Ed25519":e={name:"Ed25519"},t=s.d?["sign"]:["verify"];break;case"EdDSA":e={name:s.crv},t=s.d?["sign"]:["verify"];break;case"ECDH-ES":case"ECDH-ES+A128KW":case"ECDH-ES+A192KW":case"ECDH-ES+A256KW":e={name:s.crv},t=s.d?["deriveBits"]:[];break;default:throw new S('Invalid or unsupported JWK "alg" (Algorithm) Parameter value')}break}default:throw new S('Invalid or unsupported JWK "kty" (Key Type) Parameter value')}return{algorithm:e,keyUsages:t}}const re=async s=>{if(!s.alg)throw new TypeError('"alg" argument is required when "jwk.alg" is not present');const{algorithm:e,keyUsages:t}=Be(s),i=[e,s.ext??!1,s.key_ops??t],o={...s};return delete o.alg,delete o.use,H.subtle.importKey("jwk",o,...i)},ne=s=>E(s);let W,R;const ae=s=>(s==null?void 0:s[Symbol.toStringTag])==="KeyObject",P=async(s,e,t,i,o=!1)=>{let n=s.get(e);if(n!=null&&n[i])return n[i];const r=await re({...t,alg:i});return o&&Object.freeze(e),n?n[i]=r:s.set(e,{[i]:r}),r},ce={normalizePublicKey:(s,e)=>{if(ae(s)){let t=s.export({format:"jwk"});return delete t.d,delete t.dp,delete t.dq,delete t.p,delete t.q,delete t.qi,t.k?ne(t.k):(R||(R=new WeakMap),P(R,s,t,e))}return k(s)?s.k?E(s.k):(R||(R=new WeakMap),P(R,s,s,e,!0)):s},normalizePrivateKey:(s,e)=>{if(ae(s)){let t=s.export({format:"jwk"});return t.k?ne(t.k):(W||(W=new WeakMap),P(W,s,t,e))}return k(s)?s.k?E(s.k):(W||(W=new WeakMap),P(W,s,s,e,!0)):s}};async function Le(s,e){if(!C(s))throw new TypeError("JWK must be an object");switch(e||(e=s.alg),s.kty){case"oct":if(typeof s.k!="string"||!s.k)throw new TypeError('missing "k" (Key Value) Parameter value');return E(s.k);case"RSA":if("oth"in s&&s.oth!==void 0)throw new S('RSA JWK "oth" (Other Primes Info) Parameter value is not supported');case"EC":case"OKP":return re({...s,alg:e});default:throw new S('Unsupported "kty" (Key Type) Parameter value')}}const M=s=>s==null?void 0:s[Symbol.toStringTag],F=(s,e,t)=>{var i,o;if(e.use!==void 0&&e.use!=="sig")throw new TypeError("Invalid key for this operation, when present its use must be sig");if(e.key_ops!==void 0&&((o=(i=e.key_ops).includes)==null?void 0:o.call(i,t))!==!0)throw new TypeError(`Invalid key for this operation, when present its key_ops must include ${t}`);if(e.alg!==void 0&&e.alg!==s)throw new TypeError(`Invalid key for this operation, when present its alg must be ${s}`);return!0},Fe=(s,e,t,i)=>{if(!(e instanceof Uint8Array)){if(i&&k(e)){if(Ne(e)&&F(s,e,t))return;throw new TypeError('JSON Web Key for symmetric algorithms must have JWK "kty" (Key Type) equal to "oct" and the JWK "k" (Key Value) present')}if(!se(e))throw new TypeError(te(s,e,...J,"Uint8Array",i?"JSON Web Key":null));if(e.type!=="secret")throw new TypeError(`${M(e)} instances for symmetric algorithms must be of type "secret"`)}},Ue=(s,e,t,i)=>{if(i&&k(e))switch(t){case"sign":if(Je(e)&&F(s,e,t))return;throw new TypeError("JSON Web Key for this operation be a private JWK");case"verify":if(Pe(e)&&F(s,e,t))return;throw new TypeError("JSON Web Key for this operation be a public JWK")}if(!se(e))throw new TypeError(te(s,e,...J,i?"JSON Web Key":null));if(e.type==="secret")throw new TypeError(`${M(e)} instances for asymmetric algorithms must not be of type "secret"`);if(t==="sign"&&e.type==="public")throw new TypeError(`${M(e)} instances for asymmetric algorithm signing must be of type "private"`);if(t==="decrypt"&&e.type==="public")throw new TypeError(`${M(e)} instances for asymmetric algorithm decryption must be of type "private"`);if(e.algorithm&&t==="verify"&&e.type==="private")throw new TypeError(`${M(e)} instances for asymmetric algorithm verifying must be of type "public"`);if(e.algorithm&&t==="encrypt"&&e.type==="private")throw new TypeError(`${M(e)} instances for asymmetric algorithm encryption must be of type "public"`)};function de(s,e,t,i){e.startsWith("HS")||e==="dir"||e.startsWith("PBES2")||/^A\d{3}(?:GCM)?KW$/.test(e)?Fe(e,t,i,s):Ue(e,t,i,s)}de.bind(void 0,!1);const U=de.bind(void 0,!0);function le(s,e,t,i,o){if(o.crit!==void 0&&(i==null?void 0:i.crit)===void 0)throw new s('"crit" (Critical) Header Parameter MUST be integrity protected');if(!i||i.crit===void 0)return new Set;if(!Array.isArray(i.crit)||i.crit.length===0||i.crit.some(r=>typeof r!="string"||r.length===0))throw new s('"crit" (Critical) Header Parameter MUST be an array of non-empty strings when present');let n;t!==void 0?n=new Map([...Object.entries(t),...e.entries()]):n=e;for(const r of i.crit){if(!n.has(r))throw new S(`Extension Header Parameter "${r}" is not recognized`);if(o[r]===void 0)throw new s(`Extension Header Parameter "${r}" is missing`);if(n.get(r)&&i[r]===void 0)throw new s(`Extension Header Parameter "${r}" MUST be integrity protected`)}return new Set(i.crit)}const ze=(s,e)=>{if(e!==void 0&&(!Array.isArray(e)||e.some(t=>typeof t!="string")))throw new TypeError(`"${s}" option must be an array of strings`);if(e)return new Set(e)};function he(s,e){const t=`SHA-${s.slice(-3)}`;switch(s){case"HS256":case"HS384":case"HS512":return{hash:t,name:"HMAC"};case"PS256":case"PS384":case"PS512":return{hash:t,name:"RSA-PSS",saltLength:s.slice(-3)>>3};case"RS256":case"RS384":case"RS512":return{hash:t,name:"RSASSA-PKCS1-v1_5"};case"ES256":case"ES384":case"ES512":return{hash:t,name:"ECDSA",namedCurve:e.namedCurve};case"Ed25519":return{name:"Ed25519"};case"EdDSA":return{name:e.name};default:throw new S(`alg ${s} is not supported either by JOSE or your javascript runtime`)}}async function me(s,e,t){if(t==="sign"&&(e=await ce.normalizePrivateKey(e,s)),t==="verify"&&(e=await ce.normalizePublicKey(e,s)),G(e))return He(e,s,t),e;if(e instanceof Uint8Array){if(!s.startsWith("HS"))throw new TypeError(ee(e,...J));return H.subtle.importKey("raw",e,{hash:`SHA-${s.slice(-3)}`,name:"HMAC"},!1,[t])}throw new TypeError(ee(e,...J,"Uint8Array","JSON Web Key"))}const Ve=async(s,e,t,i)=>{const o=await me(s,e,"verify");oe(s,o);const n=he(s,o.algorithm);try{return await H.subtle.verify(n,o,t,i)}catch{return!1}};async function qe(s,e,t){if(!C(s))throw new u("Flattened JWS must be an object");if(s.protected===void 0&&s.header===void 0)throw new u('Flattened JWS must have either of the "protected" or "header" members');if(s.protected!==void 0&&typeof s.protected!="string")throw new u("JWS Protected Header incorrect type");if(s.payload===void 0)throw new u("JWS Payload missing");if(typeof s.signature!="string")throw new u("JWS Signature missing or incorrect type");if(s.header!==void 0&&!C(s.header))throw new u("JWS Unprotected Header incorrect type");let i={};if(s.protected)try{const p=E(s.protected);i=JSON.parse(I.decode(p))}catch{throw new u("JWS Protected Header is invalid")}if(!ie(i,s.header))throw new u("JWS Protected and JWS Unprotected Header Parameter names must be disjoint");const o={...i,...s.header},n=le(u,new Map([["b64",!0]]),t==null?void 0:t.crit,i,o);let r=!0;if(n.has("b64")&&(r=i.b64,typeof r!="boolean"))throw new u('The "b64" (base64url-encode payload) Header Parameter must be a boolean');const{alg:d}=o;if(typeof d!="string"||!d)throw new u('JWS "alg" (Algorithm) Header Parameter missing or invalid');const l=t&&ze("algorithms",t.algorithms);if(l&&!l.has(d))throw new X('"alg" (Algorithm) Header Parameter value not allowed');if(r){if(typeof s.payload!="string")throw new u("JWS Payload must be a string")}else if(typeof s.payload!="string"&&!(s.payload instanceof Uint8Array))throw new u("JWS Payload must be a string or an Uint8Array instance");let c=!1;typeof e=="function"?(e=await e(i,s),c=!0,U(d,e,"verify"),k(e)&&(e=await Le(e,d))):U(d,e,"verify");const f=j(w.encode(s.protected??""),w.encode("."),typeof s.payload=="string"?w.encode(s.payload):s.payload);let g;try{g=E(s.signature)}catch{throw new u("Failed to base64url decode the signature")}if(!await Ve(d,e,g,f))throw new Z;let m;if(r)try{m=E(s.payload)}catch{throw new u("Failed to base64url decode the payload")}else typeof s.payload=="string"?m=w.encode(s.payload):m=s.payload;const a={payload:m};return s.protected!==void 0&&(a.protectedHeader=i),s.header!==void 0&&(a.unprotectedHeader=s.header),c?{...a,key:e}:a}async function Ye(s,e,t){if(s instanceof Uint8Array&&(s=I.decode(s)),typeof s!="string")throw new u("Compact JWS must be a string or Uint8Array");const{0:i,1:o,2:n,length:r}=s.split(".");if(r!==3)throw new u("Invalid Compact JWS");const d=await qe({payload:o,protected:i,signature:n},e,t),l={payload:d.payload,protectedHeader:d.protectedHeader};return typeof e=="function"?{...l,key:d.key}:l}const _=s=>Math.floor(s.getTime()/1e3),pe=60,ue=pe*60,z=ue*24,Ge=z*7,je=z*365.25,Xe=/^(\+|\-)? ?(\d+|\d+\.\d+) ?(seconds?|secs?|s|minutes?|mins?|m|hours?|hrs?|h|days?|d|weeks?|w|years?|yrs?|y)(?: (ago|from now))?$/i,D=s=>{const e=Xe.exec(s);if(!e||e[4]&&e[1])throw new TypeError("Invalid time period format");const t=parseFloat(e[2]),i=e[3].toLowerCase();let o;switch(i){case"sec":case"secs":case"second":case"seconds":case"s":o=Math.round(t);break;case"minute":case"minutes":case"min":case"mins":case"m":o=Math.round(t*pe);break;case"hour":case"hours":case"hr":case"hrs":case"h":o=Math.round(t*ue);break;case"day":case"days":case"d":o=Math.round(t*z);break;case"week":case"weeks":case"w":o=Math.round(t*Ge);break;default:o=Math.round(t*je);break}return e[1]==="-"||e[4]==="ago"?-o:o},ge=s=>s.toLowerCase().replace(/^application\//,""),Ze=(s,e)=>typeof s=="string"?e.includes(s):Array.isArray(s)?e.some(Set.prototype.has.bind(new Set(s))):!1,Qe=(s,e,t={})=>{let i;try{i=JSON.parse(I.decode(e))}catch{}if(!C(i))throw new K("JWT Claims Set must be a top-level JSON object");const{typ:o}=t;if(o&&(typeof s.typ!="string"||ge(s.typ)!==ge(o)))throw new b('unexpected "typ" JWT header value',i,"typ","check_failed");const{requiredClaims:n=[],issuer:r,subject:d,audience:l,maxTokenAge:c}=t,f=[...n];c!==void 0&&f.push("iat"),l!==void 0&&f.push("aud"),d!==void 0&&f.push("sub"),r!==void 0&&f.push("iss");for(const a of new Set(f.reverse()))if(!(a in i))throw new b(`missing required "${a}" claim`,i,a,"missing");if(r&&!(Array.isArray(r)?r:[r]).includes(i.iss))throw new b('unexpected "iss" claim value',i,"iss","check_failed");if(d&&i.sub!==d)throw new b('unexpected "sub" claim value',i,"sub","check_failed");if(l&&!Ze(i.aud,typeof l=="string"?[l]:l))throw new b('unexpected "aud" claim value',i,"aud","check_failed");let g;switch(typeof t.clockTolerance){case"string":g=D(t.clockTolerance);break;case"number":g=t.clockTolerance;break;case"undefined":g=0;break;default:throw new TypeError("Invalid clockTolerance option type")}const{currentDate:h}=t,m=_(h||new Date);if((i.iat!==void 0||c)&&typeof i.iat!="number")throw new b('"iat" claim must be a number',i,"iat","invalid");if(i.nbf!==void 0){if(typeof i.nbf!="number")throw new b('"nbf" claim must be a number',i,"nbf","invalid");if(i.nbf>m+g)throw new b('"nbf" claim timestamp check failed',i,"nbf","check_failed")}if(i.exp!==void 0){if(typeof i.exp!="number")throw new b('"exp" claim must be a number',i,"exp","invalid");if(i.exp<=m-g)throw new B('"exp" claim timestamp check failed',i,"exp","check_failed")}if(c){const a=m-i.iat,p=typeof c=="number"?c:D(c);if(a-g>p)throw new B('"iat" claim timestamp check failed (too far in the past)',i,"iat","check_failed");if(a<0-g)throw new b('"iat" claim timestamp check failed (it should be in the past)',i,"iat","check_failed")}return i};async function et(s,e,t){var r;const i=await Ye(s,e,t);if((r=i.protectedHeader.crit)!=null&&r.includes("b64")&&i.protectedHeader.b64===!1)throw new K("JWTs MUST NOT use unencoded payload");const n={payload:Qe(i.protectedHeader,i.payload,t),protectedHeader:i.protectedHeader};return typeof e=="function"?{...n,key:i.key}:n}const tt=async(s,e,t)=>{const i=await me(s,e,"sign");oe(s,i);const o=await H.subtle.sign(he(s,i.algorithm),i,t);return new Uint8Array(o)};class st{constructor(e){if(!(e instanceof Uint8Array))throw new TypeError("payload must be an instance of Uint8Array");this._payload=e}setProtectedHeader(e){if(this._protectedHeader)throw new TypeError("setProtectedHeader can only be called once");return this._protectedHeader=e,this}setUnprotectedHeader(e){if(this._unprotectedHeader)throw new TypeError("setUnprotectedHeader can only be called once");return this._unprotectedHeader=e,this}async sign(e,t){if(!this._protectedHeader&&!this._unprotectedHeader)throw new u("either setProtectedHeader or setUnprotectedHeader must be called before #sign()");if(!ie(this._protectedHeader,this._unprotectedHeader))throw new u("JWS Protected and JWS Unprotected Header Parameter names must be disjoint");const i={...this._protectedHeader,...this._unprotectedHeader},o=le(u,new Map([["b64",!0]]),t==null?void 0:t.crit,this._protectedHeader,i);let n=!0;if(o.has("b64")&&(n=this._protectedHeader.b64,typeof n!="boolean"))throw new u('The "b64" (base64url-encode payload) Header Parameter must be a boolean');const{alg:r}=i;if(typeof r!="string"||!r)throw new u('JWS "alg" (Algorithm) Header Parameter missing or invalid');U(r,e,"sign");let d=this._payload;n&&(d=w.encode(N(d)));let l;this._protectedHeader?l=w.encode(N(JSON.stringify(this._protectedHeader))):l=w.encode("");const c=j(l,w.encode("."),d),f=await tt(r,e,c),g={signature:N(f),payload:""};return n&&(g.payload=I.decode(d)),this._unprotectedHeader&&(g.header=this._unprotectedHeader),this._protectedHeader&&(g.protected=I.decode(l)),g}}class it{constructor(e){this._flattened=new st(e)}setProtectedHeader(e){return this._flattened.setProtectedHeader(e),this}async sign(e,t){const i=await this._flattened.sign(e,t);if(i.payload===void 0)throw new TypeError("use the flattened module for creating JWS with b64: false");return`${i.protected}.${i.payload}.${i.signature}`}}function T(s,e){if(!Number.isFinite(e))throw new TypeError(`Invalid ${s} input`);return e}class ot{constructor(e={}){if(!C(e))throw new TypeError("JWT Claims Set MUST be an object");this._payload=e}setIssuer(e){return this._payload={...this._payload,iss:e},this}setSubject(e){return this._payload={...this._payload,sub:e},this}setAudience(e){return this._payload={...this._payload,aud:e},this}setJti(e){return this._payload={...this._payload,jti:e},this}setNotBefore(e){return typeof e=="number"?this._payload={...this._payload,nbf:T("setNotBefore",e)}:e instanceof Date?this._payload={...this._payload,nbf:T("setNotBefore",_(e))}:this._payload={...this._payload,nbf:_(new Date)+D(e)},this}setExpirationTime(e){return typeof e=="number"?this._payload={...this._payload,exp:T("setExpirationTime",e)}:e instanceof Date?this._payload={...this._payload,exp:T("setExpirationTime",_(e))}:this._payload={...this._payload,exp:_(new Date)+D(e)},this}setIssuedAt(e){return typeof e>"u"?this._payload={...this._payload,iat:_(new Date)}:e instanceof Date?this._payload={...this._payload,iat:T("setIssuedAt",_(e))}:typeof e=="string"?this._payload={...this._payload,iat:T("setIssuedAt",_(new Date)+D(e))}:this._payload={...this._payload,iat:T("setIssuedAt",e)},this}}class rt extends ot{setProtectedHeader(e){return this._protectedHeader=e,this}async sign(e,t){var o;const i=new it(w.encode(JSON.stringify(this._payload)));if(i.setProtectedHeader(this._protectedHeader),Array.isArray((o=this._protectedHeader)==null?void 0:o.crit)&&this._protectedHeader.crit.includes("b64")&&this._protectedHeader.b64===!1)throw new K("JWTs MUST NOT use unencoded payload");return i.sign(e,t)}}class nt{constructor(e){const n="http://localhost:8080/api";this.baseUrl=e||void 0||n}async getWidgetByDomain(e){const t=await fetch(`${this.baseUrl}/public/chat/widgets/domain/${e}`);if(!t.ok)throw new Error(`Failed to get widget: ${t.statusText}`);return t.json()}async createSessionToken(e,t){var f;const i=localStorage.getItem("chat_session_token");if(i)try{const{chatSessionToken:g,sessionId:h}=JSON.parse(i);if(await this.verifySessionToken(e,g))return{chatSessionToken:g,sessionId:h}}catch{localStorage.removeItem("chat_session_token")}const o=Date.now(),n=(((f=t.visitor_info)==null?void 0:f.fingerprint)||"anon")+"_"+o,r={session_id:n,widget_id:e,visitor_name:t.visitor_name,visitor_email:t.visitor_email,visitor_info:t.visitor_info,timestamp:Date.now(),iat:Math.floor(Date.now()/1e3),exp:Math.floor(Date.now()/1e3)+24*60*60},d=new TextEncoder().encode(e),c={chatSessionToken:await new rt(r).setProtectedHeader({alg:"HS256"}).setIssuedAt().setExpirationTime("24h").sign(d),sessionId:n};return localStorage.setItem("chat_session_token",JSON.stringify(c)),c}async verifySessionToken(e,t){try{const i=new TextEncoder().encode(e);return await et(t,i),!0}catch{return!1}}async initiateChat(e,t){const i=await this.createSessionToken(e,t);return console.log("session:",i),{session_token:i.chatSessionToken,session_id:i.sessionId}}async markMessagesAsRead(e){const t=await fetch(`${this.baseUrl}/public/chat/sessions/${e}/read`,{method:"POST"});if(!t.ok)throw new Error(`Failed to mark messages as read: ${t.statusText}`)}getWebSocketUrl(e,t){return`${this.baseUrl.replace("http","ws")}/public/chat/ws/widgets/${t}/chat/${e}`}}const x={SESSION:"tms_chat_session",MESSAGES:"tms_chat_messages",VISITOR_INFO:"tms_visitor_info",WIDGET_STATE:"tms_widget_state"};class at{constructor(e){this.storagePrefix=`tms_${e}_`}getKey(e){return`${this.storagePrefix}${e}`}isStorageAvailable(){try{const e="__tms_storage_test__";return localStorage.setItem(e,"test"),localStorage.removeItem(e),!0}catch{return!1}}saveSession(e){if(this.isStorageAvailable())try{const t={...e,last_activity:new Date().toISOString()};localStorage.setItem(this.getKey(x.SESSION),JSON.stringify(t))}catch(t){console.warn("Failed to save session:",t)}}getSession(){if(!this.isStorageAvailable())return null;try{const e=localStorage.getItem(this.getKey(x.SESSION));return e?JSON.parse(e):null}catch(e){return console.warn("Failed to get session:",e),null}}clearSession(){if(this.isStorageAvailable())try{localStorage.removeItem(this.getKey(x.SESSION)),localStorage.removeItem(this.getKey(x.MESSAGES))}catch(e){console.warn("Failed to clear session:",e)}}updateSessionActivity(){const e=this.getSession();e&&(e.last_activity=new Date().toISOString(),this.saveSession(e))}saveMessages(e){if(this.isStorageAvailable())try{const t=e.slice(-50);localStorage.setItem(this.getKey(x.MESSAGES),JSON.stringify(t))}catch(t){console.warn("Failed to save messages:",t)}}getMessages(){if(!this.isStorageAvailable())return[];try{const e=localStorage.getItem(this.getKey(x.MESSAGES));return e?JSON.parse(e):[]}catch(e){return console.warn("Failed to get messages:",e),[]}}addMessage(e){const t=this.getMessages();t.push(e),this.saveMessages(t)}saveVisitorInfo(e){if(this.isStorageAvailable())try{localStorage.setItem(this.getKey(x.VISITOR_INFO),JSON.stringify(e))}catch(t){console.warn("Failed to save visitor info:",t)}}getVisitorInfo(){if(!this.isStorageAvailable())return null;try{const e=localStorage.getItem(this.getKey(x.VISITOR_INFO));return e?JSON.parse(e):null}catch(e){return console.warn("Failed to get visitor info:",e),null}}saveWidgetState(e){if(this.isStorageAvailable())try{localStorage.setItem(this.getKey(x.WIDGET_STATE),JSON.stringify(e))}catch(t){console.warn("Failed to save widget state:",t)}}getWidgetState(){if(!this.isStorageAvailable())return null;try{const e=localStorage.getItem(this.getKey(x.WIDGET_STATE));return e?JSON.parse(e):null}catch(e){return console.warn("Failed to get widget state:",e),null}}cleanup(){if(this.isStorageAvailable())try{const e=[];for(let t=0;t<localStorage.length;t++){const i=localStorage.key(t);i&&i.startsWith(this.storagePrefix)&&e.push(i)}e.forEach(t=>localStorage.removeItem(t))}catch(e){console.warn("Failed to cleanup storage:",e)}}importData(e){e.session&&this.saveSession(e.session),e.messages&&this.saveMessages(e.messages),e.visitorInfo&&this.saveVisitorInfo(e.visitorInfo),e.widgetState&&this.saveWidgetState(e.widgetState)}}function fe(){return new Promise(s=>{const t=(()=>{const l=document.createElement("canvas"),c=l.getContext("2d");let f="";return c&&(l.width=200,l.height=50,c.textBaseline="top",c.font="14px Arial",c.fillStyle="#f60",c.fillRect(125,1,62,20),c.fillStyle="#069",c.fillText("🌍 Hello, world! 123",2,15),c.fillStyle="#f00",c.fillText("Canvas fingerprint",2,30),f=c.getImageData(0,0,l.width,l.height).data.slice(0,100).join("")),[navigator.userAgent,navigator.language,JSON.stringify(navigator.languages||[]),screen.width,screen.height,screen.availWidth,screen.availHeight,screen.colorDepth,screen.pixelDepth,new Date().getTimezoneOffset(),Intl.DateTimeFormat().resolvedOptions().timeZone||"",navigator.platform,navigator.cookieEnabled,navigator.doNotTrack||"",navigator.maxTouchPoints||0,navigator.hardwareConcurrency||0,window.devicePixelRatio||1,(()=>{try{const h=new(window.AudioContext||window.webkitAudioContext),m=h.createOscillator(),a=h.createAnalyser(),p=h.createGain(),_e=h.createScriptProcessor(4096,1,1);p.gain.value=0,m.frequency.value=1e3,m.type="triangle",m.connect(a),a.connect(_e),_e.connect(p),p.connect(h.destination),m.start(0);const Ee=new Uint8Array(a.frequencyBinCount);return a.getByteFrequencyData(Ee),m.stop(),h.close(),Ee.slice(0,30).join("")}catch{return"no-audio"}})(),f].join("###")})(),i=l=>{const c=[(h,m)=>{let a=m;for(let p=0;p<h.length;p++)a=(a<<5)+a+h.charCodeAt(p),a=a&4294967295;return Math.abs(a).toString(36)},(h,m)=>{let a=m;for(let p=0;p<h.length;p++)a^=h.charCodeAt(p),a=a*16777619&4294967295;return Math.abs(a).toString(36)},(h,m)=>{let a=m;for(let p=0;p<h.length;p++)a=h.charCodeAt(p)+(a<<6)+(a<<16)-a,a=a&4294967295;return Math.abs(a).toString(36)},(h,m)=>{let a=m;for(let p=0;p<h.length;p++)a=a+h.charCodeAt(p)&4294967295,a=a+(a<<10)&4294967295,a=(a^a>>>6)&4294967295;return a=a+(a<<3)&4294967295,a=(a^a>>>11)&4294967295,a=a+(a<<15)&4294967295,Math.abs(a).toString(36)}],f=[2654435769,2246822507,3266489909,668265263,374761393,3550635116,4251993797,3042594569];let g="";return f.forEach((h,m)=>{const a=c[m%c.length],p=a(l+m.toString(),h).padStart(8,"0").slice(-8);g+=p}),g},o=i(t),n=Math.floor(Date.now()/(1e3*60*60)),r=i(t+n.toString()).slice(0,16),d=o+r;s(d.slice(0,64))})}function ct(s){var e;if(!(s!=null&&s.enabled))return!0;try{const t=new Date,o=["sun","mon","tue","wed","thu","fri","sat"][t.getDay()],n=t.toTimeString().slice(0,5),r=(e=s.schedule)==null?void 0:e[o];return r!=null&&r.enabled?n>=r.open&&n<=r.close:!1}catch{return!0}}const ye={rounded:{name:"Rounded",shape:"rounded",description:"Friendly and approachable with soft rounded corners",preview:"🔵 Modern & friendly",borderRadius:"16px",shadow:"0 8px 32px rgba(0, 0, 0, 0.12)",animation:"smooth",layout:"standard"},square:{name:"Square",shape:"square",description:"Clean and professional with sharp edges",preview:"⬛ Professional & clean",borderRadius:"4px",shadow:"0 4px 20px rgba(0, 0, 0, 0.15)",animation:"fade",layout:"standard"},minimal:{name:"Minimal",shape:"minimal",description:"Ultra-clean design with minimal visual elements",preview:"⚪ Simple & clean",borderRadius:"8px",shadow:"0 2px 16px rgba(0, 0, 0, 0.08)",animation:"fade",layout:"compact"},professional:{name:"Professional",shape:"professional",description:"Enterprise-grade appearance for business use",preview:"🏢 Enterprise & formal",borderRadius:"6px",shadow:"0 6px 24px rgba(0, 0, 0, 0.1)",animation:"slide",layout:"spacious"},modern:{name:"Modern",shape:"modern",description:"Contemporary design with subtle gradients",preview:"✨ Contemporary & sleek",borderRadius:"12px",shadow:"0 10px 40px rgba(0, 0, 0, 0.15)",animation:"bounce",layout:"standard"},classic:{name:"Classic",shape:"classic",description:"Traditional chat widget with timeless design",preview:"📝 Traditional & reliable",borderRadius:"20px",shadow:"0 5px 25px rgba(0, 0, 0, 0.2)",animation:"smooth",layout:"standard"}},be={small:{width:"300px",height:"400px"},medium:{width:"350px",height:"500px"},large:{width:"400px",height:"600px"}},we={smooth:{transition:"all 0.3s cubic-bezier(0.4, 0, 0.2, 1)",transform:"translateY(0)",entry:"translateY(20px)",exit:"translateY(100%)"},bounce:{transition:"all 0.5s cubic-bezier(0.68, -0.55, 0.265, 1.55)",transform:"scale(1)",entry:"scale(0.8)",exit:"scale(0.8) translateY(100%)"},fade:{transition:"all 0.2s ease-in-out",transform:"opacity(1)",entry:"opacity(0)",exit:"opacity(0)"},slide:{transition:"all 0.4s ease-out",transform:"translateX(0)",entry:"translateX(100%)",exit:"translateX(100%)"}},ve={modern:{borderRadius:"18px 18px 4px 18px",padding:"12px 16px",maxWidth:"75%",wordBreak:"break-word",lineHeight:"1.4"},classic:{borderRadius:"20px",padding:"10px 14px",maxWidth:"70%",wordBreak:"break-word",lineHeight:"1.5"},minimal:{borderRadius:"8px",padding:"8px 12px",maxWidth:"80%",wordBreak:"break-word",lineHeight:"1.3"},rounded:{borderRadius:"25px",padding:"12px 18px",maxWidth:"75%",wordBreak:"break-word",lineHeight:"1.4"}};function dt(s){return ye[s.widget_shape]||ye.rounded}function V(s){const e=/^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(s);if(!e)return"59, 130, 246";const t=parseInt(e[1],16),i=parseInt(e[2],16),o=parseInt(e[3],16);return`${t}, ${i}, ${o}`}function lt(s){const e=/^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(s);return e?{r:parseInt(e[1],16),g:parseInt(e[2],16),b:parseInt(e[3],16)}:{r:59,g:130,b:246}}function ht(s,e,t){const[i,o,n]=[s,e,t].map(r=>(r=r/255,r<=.03928?r/12.92:Math.pow((r+.055)/1.055,2.4)));return .2126*i+.7152*o+.0722*n}function mt(s,e){const t=lt(s);return ht(t.r,t.g,t.b)<.5?`color-mix(in srgb, ${e} 60%, #ffffff 40%)`:`color-mix(in srgb, ${e} 65%, #000000 35%)`}function pt(s){const e=dt(s),t=be[s.widget_size]||be.medium,i=we[s.animation_style]||we.smooth,o=ve[s.chat_bubble_style]||ve.modern,n=mt(s.background_color||"#ffffff",s.secondary_color||"#6b7280");return`
    :root {
      --tms-primary-color: ${s.primary_color};
      --tms-primary-color-rgb: ${V(s.primary_color)};
      --tms-secondary-color: ${s.secondary_color||"#6b7280"};
      --tms-secondary-color-rgb: ${V(s.secondary_color||"#6b7280")};
      --tms-background-color: ${s.background_color||"#ffffff"};
      --tms-background-color-rgb: ${V(s.background_color||"#ffffff")};
      --tms-placeholder-color: ${n};
      --tms-widget-width: ${t.width};
      --tms-widget-height: ${t.height};
      --tms-border-radius: ${e.borderRadius};
      --tms-shadow: ${e.shadow};
      --tms-animation: ${i.transition};
      --tms-bubble-border-radius: ${o.borderRadius};
      --tms-bubble-padding: ${o.padding};
      --tms-bubble-max-width: ${o.maxWidth};
    }

    /* Main Widget Container */
    .tms-widget-container {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
      position: fixed;
      ${s.position==="bottom-right"?"right: 24px;":"left: 24px;"}
      bottom: 96px;
      width: var(--tms-widget-width);
      height: var(--tms-widget-height);
      z-index: 2147483647;
      border-radius: 16px;
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25), 0 0 0 1px rgba(255, 255, 255, 0.05);
      overflow: hidden;
      background: #ffffff;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      display: none;
      flex-direction: column;
      backdrop-filter: blur(20px);
      background: var(--tms-background-color);
    }

    .tms-widget-container.open {
      display: flex;
      opacity: 1;
      transform: translateY(0) scale(1);
    }

    .tms-widget-container.opening {
      animation: tms-widget-enter 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
    }

    .tms-widget-container.closing {
      animation: tms-widget-exit 0.3s cubic-bezier(0.4, 0, 1, 1) forwards;
    }

    @keyframes tms-widget-enter {
      0% {
        opacity: 0;
        transform: translateY(20px) scale(0.9);
      }
      100% {
        opacity: 1;
        transform: translateY(0) scale(1);
      }
    }

    @keyframes tms-widget-exit {
      0% {
        opacity: 1;
        transform: translateY(0) scale(1);
      }
      100% {
        opacity: 0;
        transform: translateY(20px) scale(0.95);
      }
    }

    /* Header Section */
    .tms-chat-header {
      background: linear-gradient(135deg, var(--tms-primary-color) 0%, color-mix(in srgb, var(--tms-primary-color) 85%, #000) 100%);
      color: white;
      padding: 20px 20px 18px 20px;
      display: flex;
      justify-content: space-between;
      align-items: center;
      position: relative;
      box-shadow: 0 4px 12px rgba(var(--tms-primary-color-rgb), 0.15);
    }

    .tms-chat-header::after {
      content: '';
      position: absolute;
      bottom: 0;
      left: 0;
      right: 0;
      height: 1px;
      background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
    }

    .tms-agent-info {
      display: flex;
      align-items: center;
      gap: 12px;
      flex: 1;
    }

    .tms-agent-avatar {
      width: 40px;
      height: 40px;
      border-radius: 50%;
      background: rgba(255, 255, 255, 0.15);
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 600;
      font-size: 16px;
      color: white;
      border: 2px solid rgba(255, 255, 255, 0.2);
      position: relative;
      overflow: hidden;
    }

    .tms-agent-avatar::before {
      content: '';
      position: absolute;
      top: -50%;
      left: -50%;
      width: 200%;
      height: 200%;
      background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.1), transparent);
      transform: rotate(45deg);
      animation: avatar-shine 3s infinite;
    }

    @keyframes avatar-shine {
      0%, 100% { transform: translateX(-100%) translateY(-100%) rotate(45deg); }
      50% { transform: translateX(100%) translateY(100%) rotate(45deg); }
    }

    .tms-agent-avatar img {
      width: 100%;
      height: 100%;
      border-radius: 50%;
      object-fit: cover;
      position: relative;
      z-index: 1;
    }

    .tms-agent-details {
      flex: 1;
      min-width: 0;
    }

    .tms-agent-name {
      font-weight: 600;
      font-size: 16px;
      margin: 0 0 2px 0;
      color: white;
      line-height: 1.2;
    }

    .tms-agent-status {
      font-size: 13px;
      opacity: 0.9;
      color: rgba(255, 255, 255, 0.8);
      display: flex;
      align-items: center;
      gap: 6px;
    }

    .tms-status-indicator {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #10b981;
      animation: pulse 2s infinite;
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.5; }
    }

    .tms-header-close {
      background: rgba(255, 255, 255, 0.1);
      border: none;
      color: white;
      width: 32px;
      height: 32px;
      border-radius: 8px;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 18px;
      font-weight: 300;
      transition: all 0.2s ease;
      backdrop-filter: blur(10px);
    }

    .tms-header-close:hover {
      background: rgba(255, 255, 255, 0.2);
      transform: scale(1.05);
    }

    .tms-header-close:active {
      transform: scale(0.95);
    }

    /* Messages Area */
    .tms-chat-body {
      flex: 1;
      display: flex;
      flex-direction: column;
      background: var(--tms-background-color);
      min-height: 0;
    }

    .tms-messages-container {
      flex: 1;
      overflow-y: auto;
      padding: 20px 16px 12px 16px;
      scroll-behavior: smooth;
    }

    .tms-messages-container::-webkit-scrollbar {
      width: 4px;
    }

    .tms-messages-container::-webkit-scrollbar-track {
      background: transparent;
    }

    .tms-messages-container::-webkit-scrollbar-thumb {
      background: rgba(0, 0, 0, 0.1);
      border-radius: 2px;
    }

    .tms-messages-container::-webkit-scrollbar-thumb:hover {
      background: rgba(0, 0, 0, 0.2);
    }

    /* Message Bubbles */
    .tms-message-wrapper {
      margin-bottom: 16px;
      display: flex;
      flex-direction: column;
    }

    .tms-message-wrapper.visitor {
      align-items: flex-end;
    }

    .tms-message-wrapper.agent {
      align-items: flex-start;
    }

    .tms-message-bubble {
      position: relative;
      border-radius: 18px;
      padding: 12px 16px;
      max-width: 280px;
      word-wrap: break-word;
      line-height: 1.4;
      font-size: 14px;
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
      animation: message-appear 0.3s ease-out;
    }

    @keyframes message-appear {
      0% {
        opacity: 0;
        transform: translateY(10px) scale(0.9);
      }
      100% {
        opacity: 1;
        transform: translateY(0) scale(1);
      }
    }

    .tms-message-bubble.visitor {
      background: linear-gradient(135deg, var(--tms-primary-color) 0%, color-mix(in srgb, var(--tms-primary-color) 90%, #000) 100%);
      color: white;
      border-bottom-right-radius: 6px;
    }

    .tms-message-bubble.agent {
      background: var(--tms-secondary-color);
      color: color-mix(in srgb, var(--tms-secondary-color) 15%, #000);
      border-bottom-left-radius: 6px;
      border: 1px solid color-mix(in srgb, var(--tms-secondary-color) 80%, #fff);
    }

    .tms-message-bubble.system {
      background: #f3f4f6;
      color: #6b7280;
      font-style: italic;
      text-align: center;
      border-radius: 12px;
      font-size: 13px;
      max-width: 100%;
    }

    .tms-message-time {
      font-size: 11px;
      opacity: 0.6;
      margin-top: 4px;
      padding: 0 4px;
    }

    .tms-message-wrapper.visitor .tms-message-time {
      text-align: right;
      color: #6b7280;
    }

    .tms-message-wrapper.agent .tms-message-time {
      text-align: left;
      color: #9ca3af;
    }

    /* Typing Indicator */
    .tms-typing-indicator {
      padding: 8px 16px;
      font-size: 13px;
      color: #6b7280;
      font-style: italic;
      min-height: 24px;
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .tms-typing-dots {
      display: flex;
      gap: 2px;
    }

    .tms-typing-dot {
      width: 4px;
      height: 4px;
      border-radius: 50%;
      background: #9ca3af;
      animation: typing-bounce 1.4s infinite ease-in-out;
    }

    .tms-typing-dot:nth-child(1) { animation-delay: -0.32s; }
    .tms-typing-dot:nth-child(2) { animation-delay: -0.16s; }

    @keyframes typing-bounce {
      0%, 80%, 100% { 
        transform: scale(0.8);
        opacity: 0.5;
      }
      40% { 
        transform: scale(1);
        opacity: 1;
      }
    }

    /* Input Area */
    .tms-input-area {
      background: transparent;
      border-top: 1px solid rgba(var(--tms-secondary-color-rgb), 0.2);
      display: flex;
      flex-direction: column;
      gap: 8px;
      min-height: 100px; /* Ensure adequate space for input */
  /* Keep input area anchored at the bottom while header stays static */
  margin-top: auto;
    }

    .tms-input-controls {
      display: flex;
      align-items: center;
      justify-content: space-between;
      position: relative; /* allow absolute centering of reactions */
  width: 100%;
    }

    .tms-input-wrapper {
      flex: 1;
      display: flex;
      background: var(--tms-background-color);
      border-top: 2px solid rgba(var(--tms-secondary-color-rgb), 0.8);
      padding: 8px 12px;
      transition: all 0.2s ease;
    }

    .tms-input-wrapper:focus-within {
      /* border: 2px solid color-mix(in srgb, var(--tms-secondary-color) 100%, var(--tms-background-color)); */
      /* Use a subtle ring mixed from background and black for better theme contrast */
      box-shadow: 0 0 0 1px color-mix(in srgb, var(--tms-background-color) 95%, #000 5%)
    }

    .tms-file-upload-btn {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: rgba(var(--tms-secondary-color-rgb), 0.1);
      border: 1px solid rgba(var(--tms-secondary-color-rgb), 0.2);
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      color: rgba(var(--tms-secondary-color-rgb), 0.7);
      transition: all 0.2s ease;
      flex-shrink: 0;
    }

    .tms-file-upload-btn:hover {
      background: rgba(var(--tms-secondary-color-rgb), 0.15);
      color: var(--tms-primary-color);
      border-color: var(--tms-primary-color);
    }

    .tms-message-input {
      flex: 1;
      border: none;
      outline: none;
      background: transparent;
      font-size: 14px;
      font-family: inherit;
      /* min-height: 60px;  Larger input area  */
      max-height: 120px;
      line-height: 1.5;
      padding: 4px 0;
      overflow-y: auto;
      resize: none;
    }

    /* Enhanced placeholder visibility with complementary colors */
    .tms-editable:empty:before {
      content: attr(data-placeholder);
      color: var(--tms-placeholder-color);
      pointer-events: none;
      font-style: italic;
      opacity: 0.8;
    }

    /* Reactions */
    .tms-reaction-group { 
      display: flex; 
      gap: 6px; 
      position: absolute; /* center regardless of other controls */
      left: 50%;
      transform: translateX(-50%);
      justify-content: center;
      align-items: center;
      margin-bottom: 20px;
      pointer-events: auto;
      z-index: 2; /* ensure reactions sit above other controls */
    }
    
    .tms-thumb-button {
      width: 32px; 
      height: 32px; 
      border-radius: 50%;
      border: 1px solid rgba(var(--tms-secondary-color-rgb), 0.3);
      background: rgba(var(--tms-secondary-color-rgb), 0.08);
      cursor: pointer; 
      display: flex; 
      align-items: center; 
      justify-content: center;
      transition: all 0.2s ease; 
      flex-shrink: 0; 
      font-size: 14px;
      color: rgba(var(--tms-secondary-color-rgb), 0.6);
    }
    
    .tms-thumb-button:hover { 
      color: var(--tms-primary-color); 
      border-color: var(--tms-primary-color);
      background: rgba(var(--tms-primary-color-rgb), 0.1);
      transform: scale(1.05);
    }

    .tms-thumb-button:active {
      transform: scale(0.95);
    }
    
    /* Thumb animation */
    .tms-thumb-animate {
      animation: thumb-bounce 0.6s ease-out;
      background: var(--tms-primary-color) !important;
      color: white !important;
      border-color: var(--tms-primary-color) !important;
    }
    
    @keyframes thumb-bounce {
      0% { transform: scale(1); }
      30% { transform: scale(1.3) rotate(10deg); }
      60% { transform: scale(1.1) rotate(-5deg); }
      100% { transform: scale(1) rotate(0deg); }
    }

    /* Powered By */
    .tms-powered-by {
      text-align: center;
      padding: 8px;
      font-size: 11px;
      color: #9ca3af;
    }

    /* Visitor Info Form */
    .tms-visitor-info-form {
      padding: 20px 16px;
      background: var(--tms-background-color);
      border-bottom: 1px solid color-mix(in srgb, var(--tms-secondary-color) 20%, var(--tms-background-color));
    }

    .tms-visitor-info-title {
      font-size: 16px;
      font-weight: 600;
      color: color-mix(in srgb, var(--tms-secondary-color) 15%, #000);
      margin: 0 0 8px 0;
      text-align: center;
    }

    .tms-visitor-info-subtitle {
      font-size: 14px;
      color: color-mix(in srgb, var(--tms-secondary-color) 40%, #000);
      margin: 0 0 16px 0;
      text-align: center;
      line-height: 1.4;
    }

    .tms-visitor-form-field {
      margin-bottom: 12px;
    }

    .tms-visitor-form-label {
      display: block;
      font-size: 13px;
      font-weight: 500;
      color: color-mix(in srgb, var(--tms-secondary-color) 20%, #000);
      margin-bottom: 4px;
    }

    .tms-visitor-form-input {
      width: 100%;
      padding: 10px 12px;
      border: 2px solid color-mix(in srgb, var(--tms-secondary-color) 100%, var(--tms-background-color));
      border-radius: 8px;
      font-size: 14px;
      font-family: inherit;
      background: var(--tms-background-color);
      color: color-mix(in srgb, var(--tms-secondary-color) 20%, #000);
      transition: all 0.2s ease;
      box-sizing: border-box;
    }

    .tms-visitor-form-input:focus {
      outline: none;
      box-shadow: 0 0 0 3px rgba(var(--tms-primary-color-rgb), 0.15);
    }

    .tms-visitor-form-input::placeholder {
      color: var(--tms-placeholder-color);
      opacity: 0.8;
    }

    .tms-visitor-form-actions {
      display: flex;
      gap: 8px;
      margin-top: 16px;
    }

    .tms-visitor-form-button {
      flex: 1;
      padding: 12px 16px;
      border: none;
      border-radius: 8px;
      font-size: 14px;
      font-weight: 600;
      cursor: pointer;
      transition: all 0.2s ease;
      font-family: inherit;
    }

    .tms-visitor-form-button.primary {
      background: var(--tms-primary-color);
      color: white;
    }

    .tms-visitor-form-button.primary:hover {
      background: color-mix(in srgb, var(--tms-primary-color) 90%, #000);
      transform: translateY(-1px);
    }

    .tms-visitor-form-button.primary:active {
      transform: translateY(0);
    }

    .tms-visitor-form-button.secondary {
      background: color-mix(in srgb, var(--tms-secondary-color) 15%, var(--tms-background-color));
      color: color-mix(in srgb, var(--tms-secondary-color) 20%, #000);
      border: 1px solid color-mix(in srgb, var(--tms-secondary-color) 100%, var(--tms-background-color));
    }

    .tms-visitor-form-button.secondary:hover {
      background: color-mix(in srgb, var(--tms-secondary-color) 20%, var(--tms-background-color));
    }

    /* Toggle Button */
    .tms-toggle-button {
      position: fixed;
      ${s.position==="bottom-right"?"right: 24px;":"left: 24px;"}
      bottom: 24px;
      width: 64px;
      height: 64px;
      background: linear-gradient(135deg, var(--tms-primary-color) 0%, color-mix(in srgb, var(--tms-primary-color) 85%, #000) 100%);
      border-radius: 50%;
      cursor: pointer;
      display: flex;
      align-items: center;
      justify-content: center;
      box-shadow: 0 8px 25px rgba(var(--tms-primary-color-rgb), 0.3), 0 4px 12px rgba(0, 0, 0, 0.15);
      z-index: 2147483647;
      transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
      border: none;
      outline: none;
      color: white;
      font-size: 24px;
    }

    .tms-toggle-button:hover {
      transform: translateY(-2px) scale(1.05);
      box-shadow: 0 12px 35px rgba(var(--tms-primary-color-rgb), 0.4), 0 8px 20px rgba(0, 0, 0, 0.2);
    }

    .tms-toggle-button:active {
      transform: translateY(0) scale(1.02);
    }

    .tms-toggle-button:focus {
      outline: 3px solid rgba(var(--tms-primary-color-rgb), 0.3);
      outline-offset: 2px;
    }

    .tms-toggle-button svg {
      width: 28px;
      height: 28px;
      transition: all 0.2s ease;
    }

    .tms-toggle-button:hover svg {
      transform: scale(1.1);
    }

    /* Notification Badge */
    .tms-notification-badge {
      position: absolute;
      top: -4px;
      right: -4px;
      background: #ef4444;
      color: white;
      border-radius: 50%;
      min-width: 22px;
      height: 22px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 11px;
      font-weight: 600;
      border: 2px solid white;
      animation: badge-bounce 0.5s ease-out;
    }

    @keyframes badge-bounce {
      0% { transform: scale(0); }
      50% { transform: scale(1.2); }
      100% { transform: scale(1); }
    }

    /* External Powered By badge (outside container) */
    .tms-powered-badge {
      position: fixed;
      bottom: 96px; /* Above toggle button */
      ${s.position==="bottom-right"?"right: 24px;":"left: 24px;"}
      background: rgba(0, 0, 0, 0.7);
      color: #fff;
      font-size: 11px;
      padding: 6px 12px;
      border-radius: 999px;
      text-decoration: none;
      z-index: 2147483646;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
      backdrop-filter: blur(10px);
      transition: all 0.3s ease;
      opacity: 0.8;
    }

    .tms-powered-badge:hover {
      opacity: 1;
      background: rgba(0, 0, 0, 0.8);
    }
    
    .tms-powered-badge.open { 
      bottom: 40px; /* Dynamic height: widget height + bottom offset + padding */
      ${s.position==="bottom-right"?`right: calc((${t.width} - 50px) / 2);`:`left: calc(${t.width} - 50px);`}
    }

    /* Mobile Responsiveness */
    @media (max-width: 480px) {
      .tms-widget-container {
        ${s.position==="bottom-right"?"right: 16px;":"left: 16px;"}
        bottom: 80px;
        width: calc(100vw - 32px);
        max-width: 360px;
        height: 500px;
      }

      .tms-toggle-button {
        ${s.position==="bottom-right"?"right: 16px;":"left: 16px;"}
        bottom: 16px;
        width: 56px;
        height: 56px;
      }

      .tms-chat-header {
        padding: 16px;
      }

      .tms-agent-avatar {
        width: 36px;
        height: 36px;
      }

      .tms-message-bubble {
        max-width: 240px;
      }

      .tms-powered-badge { 
        bottom: 80px; 
        ${s.position==="bottom-right"?"right: 16px;":"left: 16px;"}
      }
      
      .tms-powered-badge.open { 
        bottom: 580px; /* Above mobile widget */
      }
    }

    /* Dark mode support */
    
    /* Custom CSS */
    ${s.custom_css||""}
  `}function xe(s){const e=document.getElementById("tms-widget-styles");e&&e.remove();const t=document.createElement("style");t.id="tms-widget-styles",t.textContent=pt(s),document.head.appendChild(t)}function q(s,e=!0){if(e)try{const t=new(window.AudioContext||window.webkitAudioContext),i=t.createOscillator(),o=t.createGain();i.connect(o),o.connect(t.destination);const n={message:[800,600],notification:[600,800],error:[300,200]},[r,d]=n[s];i.frequency.setValueAtTime(r,t.currentTime),i.frequency.setValueAtTime(d,t.currentTime+.1),o.gain.setValueAtTime(.1,t.currentTime),o.gain.exponentialRampToValueAtTime(.01,t.currentTime+.2),i.start(t.currentTime),i.stop(t.currentTime+.2)}catch(t){console.debug("Audio notification not available:",t)}}class Y{constructor(e){this.options=e,this.emitter=new O,this.widget=null,this.session=null,this.container=null,this.toggleButton=null,this.websocket=null,this.isOpen=!1,this.messages=[],this.isTyping=!1,this.typingTimeout=null,this.unreadCount=0,this.isBusinessHoursOpen=!0,this.reconnectAttempts=0,this.maxReconnectAttempts=5,this.reconnectDelay=3e3,this.isConnected=!1,this.poweredBadge=null,this.api=new nt(e.apiUrl),this.storage=new at(e.widgetId),this.init()}getBubbleStyleIcon(e){switch(e){case"modern":return`<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
        </svg>`;case"classic":return`<svg width="28" height="28" viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12c0 1.54.36 2.98.97 4.29L1 23l6.71-1.97C9.02 21.64 10.46 22 12 22c5.52 0 10-4.48 10-10S17.52 2 12 2z"/>
        </svg>`;case"minimal":return`<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
          <polyline points="22,6 12,13 2,6"/>
        </svg>`;case"bot":return`<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 8V4H8"/>
          <rect width="16" height="12" x="4" y="8" rx="2"/>
          <path d="M2 14h2"/>
          <path d="M20 14h2"/>
          <path d="M15 13v2"/>
          <path d="M9 13v2"/>
        </svg>`;default:return`<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
        </svg>`}}async init(){try{if(await this.restoreSession(),this.widget=await this.api.getWidgetByDomain(this.options.domain),!this.widget)throw new Error("Widget not found for domain");this.isBusinessHoursOpen=ct(this.widget.business_hours),xe(this.widget),this.createWidget(),this.widget.auto_open_delay>0&&!this.session&&setTimeout(()=>this.open(),this.widget.auto_open_delay*1e3)}catch(e){console.error("Failed to initialize chat widget:",e),this.emitter.emit("error","Failed to initialize chat widget")}}async restoreSession(){const e=this.storage.getSession();e&&(this.session={id:e.session_id,token:e.token,widget_id:e.widget_id,status:"active"},this.messages=this.storage.getMessages(),this.storage.updateSessionActivity())}createWidget(){if(!this.widget)return;this.container=document.createElement("div"),this.container.id="tms-chat-widget",this.container.className="tms-widget-container";const e=`
      <div class="tms-chat-header">
        <div class="tms-agent-info">
          ${this.widget.show_agent_avatars&&this.widget.agent_avatar_url?`<div class="tms-agent-avatar">
              <img src="${this.widget.agent_avatar_url}" alt="${this.widget.agent_name}" />
            </div>`:`<div class="tms-agent-avatar">${this.widget.agent_name.charAt(0).toUpperCase()}</div>`}
          <div class="tms-agent-details">
            <div class="tms-agent-name">${this.widget.agent_name}</div>
            <div class="tms-agent-status">
              <div class="tms-status-indicator"></div>
              ${this.isBusinessHoursOpen?"Online now":"Away"}
            </div>
          </div>
        </div>
        <button class="tms-header-close" aria-label="Close chat">×</button>
      </div>
    `,t=`
      <div class="tms-chat-body">
        <div id="tms-visitor-info-form" class="tms-visitor-info-form" style="display: none;">
          <div class="tms-visitor-info-title">Start a conversation</div>
          <form id="tms-visitor-form">
          ${this.widget.require_name?`
            <div class="tms-visitor-form-field">
              <label class="tms-visitor-form-label" for="tms-visitor-name">Name *</label>
              <input 
                id="tms-visitor-name" 
                type="text" 
                class="tms-visitor-form-input"
                placeholder="Enter your name"
                required
              />
            </div>`:""}
            ${this.widget.require_email?`
              <div class="tms-visitor-form-field">
                <label class="tms-visitor-form-label" for="tms-visitor-email">Email *</label>
                <input 
                  id="tms-visitor-email" 
                  type="email" 
                  class="tms-visitor-form-input"
                  placeholder="Enter your email"
                  required
                />
              </div>
            `:""}

            ${this.widget.require_name||this.widget.require_email?`<div class="tms-visitor-form-actions">
              <button type="button" id="tms-visitor-cancel" class="tms-visitor-form-button secondary">Cancel</button>
              <button type="submit" id="tms-visitor-start" class="tms-visitor-form-button primary">Start Chat</button>
            </div>`:""}
          </form>
        </div>
        
        <div class="tms-messages-container" id="tms-chat-messages"></div>
        
        <div class="tms-typing-indicator" id="tms-chat-typing" style="display: none;">
          <div class="tms-typing-dots">
            <div class="tms-typing-dot"></div>
            <div class="tms-typing-dot"></div>
            <div class="tms-typing-dot"></div>
          </div>
          <span id="tms-typing-text"></span>
        </div>
        
        <div class="tms-input-area">
          <div class="tms-input-controls">
            ${this.widget.allow_file_uploads?`
              <label class="tms-file-upload-btn" for="tms-file-input" title="Attach file">
                📎
                <input id="tms-file-input" type="file" style="display: none;" accept="image/*,.pdf,.doc,.docx,.txt" />
              </label>
            `:""}
            <div class="tms-reaction-group">
              <button id="tms-thumb-up" class="tms-thumb-button" title="Thumbs up" aria-label="Thumbs up">👍</button>
              <button id="tms-thumb-down" class="tms-thumb-button" title="Thumbs down" aria-label="Thumbs down">👎</button>
            </div>
          </div>
          <div class="tms-input-wrapper">
            <div
              id="tms-chat-input"
              class="tms-message-input tms-editable"
              contenteditable="true"
              role="textbox"
              aria-multiline="true"
              data-placeholder="Type your message and press Enter..."
            ></div>
          </div>
        </div>
      </div>
    `;if(this.container.innerHTML=e+t,this.toggleButton=document.createElement("button"),this.toggleButton.id="tms-chat-toggle",this.toggleButton.className="tms-toggle-button",this.toggleButton.setAttribute("aria-label","Open chat"),this.toggleButton.innerHTML=`
      ${this.getBubbleStyleIcon(this.widget.chat_bubble_style)}
      ${this.unreadCount>0?`
        <div class="tms-notification-badge">${this.unreadCount>9?"9+":this.unreadCount}</div>
      `:""}
    `,document.body.appendChild(this.container),document.body.appendChild(this.toggleButton),this.widget.show_powered_by){const i=document.createElement("a");i.className="tms-powered-badge",i.href="https://bareuptime.com/tms",i.target="_blank",i.rel="noopener noreferrer",i.textContent="Powered by TMS",i.style.display="none",document.body.appendChild(i),this.poweredBadge=i}this.attachEventListeners(),this.messages.length>0?this.messages.forEach(i=>this.displayMessage(i)):this.session||this.showWelcomeMessage()}showWelcomeMessage(){if(!this.widget||!this.session)return;const e=this.widget.custom_greeting||this.widget.welcome_message,t={id:"welcome-"+Date.now(),content:e,author_type:"system",author_name:this.widget.agent_name,created_at:new Date().toISOString(),message_type:"text",is_private:!1};this.displayMessage(t)}displayMessage(e){var d;const t=document.getElementById("tms-chat-messages");if(!t)return;const i=document.createElement("div");i.className=`tms-message-wrapper ${e.author_type}`;const o=e.author_type==="agent"||e.author_type==="ai-agent",n=document.createElement("div");n.className=`tms-message-bubble ${e.author_type}`,n.innerHTML=this.escapeHtml(e.content);const r=document.createElement("div");r.className="tms-message-time",r.textContent=new Date(e.created_at).toLocaleTimeString([],{hour:"2-digit",minute:"2-digit"}),i.appendChild(n),i.appendChild(r),t.appendChild(i),requestAnimationFrame(()=>{t.scrollTop=t.scrollHeight}),!this.isOpen&&o&&(this.unreadCount++,this.updateNotificationBadge()),o&&((d=this.widget)!=null&&d.sound_enabled)&&q("message",!0)}attachEventListeners(){if(!this.container||!this.toggleButton)return;const e=this.container.querySelector(".tms-header-close"),t=this.container.querySelector("#tms-chat-input"),i=this.container.querySelector("#tms-file-input"),o=this.container.querySelector("#tms-thumb-up"),n=this.container.querySelector("#tms-thumb-down"),r=this.container.querySelector("#tms-visitor-form"),d=this.container.querySelector("#tms-visitor-cancel"),l=this.container.querySelector("#tms-visitor-name");this.toggleButton.addEventListener("click",()=>this.toggle()),e==null||e.addEventListener("click",()=>this.close()),r==null||r.addEventListener("submit",c=>{c.preventDefault(),this.handleVisitorFormSubmit()}),d==null||d.addEventListener("click",()=>{this.hideVisitorForm(),this.close()}),l==null||l.addEventListener("focus",()=>{l.value||l.focus()}),t==null||t.addEventListener("keydown",c=>{c.key==="Enter"&&!c.shiftKey&&(c.preventDefault(),this.sendMessage())}),t==null||t.addEventListener("input",()=>{this.handleTyping(),this.autoResizeEditable(t)}),t==null||t.addEventListener("keyup",()=>{(!t.textContent||!t.textContent.trim())&&this.stopTyping()}),t==null||t.addEventListener("blur",()=>{this.stopTyping()}),i==null||i.addEventListener("change",c=>{const f=c.target.files;f&&f.length>0&&this.handleFileUpload(f[0])}),o==null||o.addEventListener("click",()=>{this.animateThumb(o,"👍"),this.sendQuickReaction("👍")}),n==null||n.addEventListener("click",()=>{this.animateThumb(n,"👎"),this.sendQuickReaction("👎")}),document.addEventListener("keydown",c=>{c.key==="Escape"&&this.isOpen&&this.close()})}autoResizeEditable(e){e.style.height="auto";const t=100,i=Math.min(e.scrollHeight,t);e.style.height=i+"px",e.style.overflowY=e.scrollHeight>t?"auto":"hidden"}animateThumb(e,t){e.classList.add("tms-thumb-animate");const i=e.textContent;e.textContent=t,setTimeout(()=>{e.classList.remove("tms-thumb-animate"),e.textContent=i},600)}async handleFileUpload(e){var i;if(!this.session||!((i=this.widget)!=null&&i.allow_file_uploads))return;const t=10*1024*1024;if(e.size>t){this.showError("File size must be less than 10MB");return}try{const o={id:"file-"+Date.now(),content:`📎 Uploaded: ${e.name}`,author_type:"visitor",author_name:"You",created_at:new Date().toISOString(),message_type:"file",is_private:!1};this.addMessage(o)}catch{this.showError("Failed to upload file")}}showError(e){var i;const t={id:"error-"+Date.now(),content:`⚠️ ${e}`,author_type:"system",author_name:"System",created_at:new Date().toISOString(),message_type:"text",is_private:!1};this.displayMessage(t),(i=this.widget)!=null&&i.sound_enabled&&q("error",!0)}showVisitorForm(){const e=document.getElementById("tms-visitor-info-form"),t=document.getElementById("tms-chat-messages");e&&t&&(e.style.display="block",t.style.display="none",setTimeout(()=>{const i=document.getElementById("tms-visitor-name");i==null||i.focus()},100))}hideVisitorForm(){const e=document.getElementById("tms-visitor-info-form"),t=document.getElementById("tms-chat-messages");e&&t&&(e.style.display="none",t.style.display="block")}async handleVisitorFormSubmit(){var r;console.log("1");const e=document.getElementById("tms-visitor-name"),t=document.getElementById("tms-visitor-email");if(console.log("2"),!e)return;console.log("3");const i=e.value.trim(),o=t==null?void 0:t.value.trim();if(console.log("4"),!i){e.focus();return}if(console.log("5"),(r=this.widget)!=null&&r.require_email&&!o){t==null||t.focus();return}console.log("6");const n=await fe();if(this.storage.saveVisitorInfo({name:i,email:o,fingerprint:n,last_visit:new Date().toISOString()}),console.log("7"),this.hideVisitorForm(),this.widget){const d=this.widget.custom_greeting||this.widget.welcome_message,l={id:"welcome-"+Date.now(),content:d,author_type:"system",author_name:this.widget.agent_name,created_at:new Date().toISOString(),message_type:"text",is_private:!1};this.displayMessage(l),console.log("8")}console.log("9"),await this.startChatSessionWithVisitorInfo({name:i,email:o})}async startChatSessionWithVisitorInfo(e){if(this.widget)try{const t=await fe(),i={visitor_name:e.name,visitor_email:e.email,initial_message:this.widget.welcome_message,visitor_info:{fingerprint:t,user_agent:navigator.userAgent,timezone:Intl.DateTimeFormat().resolvedOptions().timeZone,language:navigator.language}},o=await this.api.initiateChat(this.widget.id,i);this.session={id:o.session_id,token:o.session_token,widget_id:this.widget.id,status:"active",visitor_name:e.name},this.storage.saveSession({session_id:this.session.id,token:this.session.token,widget_id:this.widget.id,visitor_name:e.name,visitor_email:e.email,created_at:new Date().toISOString(),last_activity:new Date().toISOString()}),this.connectWebSocket(),this.emitter.emit("session:started",this.session)}catch(t){console.error("Failed to start chat session:",t),this.emitter.emit("error","Failed to start chat session"),this.showError("Unable to connect. Please try again."),this.showVisitorForm()}}shouldOpenVisitorForm(){var e,t,i,o;return(e=this.widget)!=null&&e.require_email&&!((t=this.storage.getVisitorInfo())!=null&&t.email)?(console.log("1"),!0):(i=this.widget)!=null&&i.require_name&&!((o=this.storage.getVisitorInfo())!=null&&o.name)?(console.log("2"),!0):(console.log("4"),!1)}async open(){if(!(!this.widget||!this.container)){if(console.log("visitor form opened"),this.container.classList.add("opening"),this.container.classList.add("open"),this.isOpen=!0,this.toggleButton&&this.toggleButton.setAttribute("aria-label","Close chat"),this.poweredBadge&&(this.poweredBadge.style.display="block",this.poweredBadge.classList.add("open")),this.session)console.log("chat session already started"),this.storage.updateSessionActivity(),this.isConnected||this.connectWebSocket();else{console.log("no chat session detected"),this.shouldOpenVisitorForm()||(console.log("no visitor form opened"),this.startChatSessionWithVisitorInfo({name:"",email:""}));const e=this.storage.getVisitorInfo();e&&e.name?await this.startChatSessionWithVisitorInfo({name:e.name,email:e.email}):(console.log("no stored visitor info found"),this.showVisitorForm())}this.unreadCount>0&&(this.unreadCount=0,this.updateNotificationBadge(),this.markUnreadMessagesAsRead()),setTimeout(()=>{const e=document.getElementById("tms-chat-input");e==null||e.focus()},300)}}markUnreadMessagesAsRead(){if(!this.isConnected)return;this.messages.filter(t=>t.author_type==="agent"||t.author_type==="ai-agent").slice(-5).forEach(t=>{this.sendReadReceipt(t.id)})}close(){this.container&&(this.container.classList.add("closing"),this.container.classList.remove("open"),this.isOpen=!1,this.toggleButton&&this.toggleButton.setAttribute("aria-label","Open chat"),this.poweredBadge&&(this.poweredBadge.classList.remove("open"),this.poweredBadge.style.display="none"),this.unreadCount=0,this.updateNotificationBadge(),this.storage.saveWidgetState({is_minimized:!1,unread_count:0,last_interaction:new Date().toISOString()}),setTimeout(()=>{var e;(e=this.container)==null||e.classList.remove("opening","closing")},300))}toggle(){this.isOpen?this.close():this.open()}updateNotificationBadge(){if(!this.toggleButton)return;const e=this.toggleButton.querySelector(".tms-notification-badge");if(this.unreadCount>0)if(e)e.textContent=this.unreadCount>9?"9+":this.unreadCount.toString();else{const t=document.createElement("div");t.className="tms-notification-badge",t.textContent=this.unreadCount>9?"9+":this.unreadCount.toString(),this.toggleButton.appendChild(t)}else e&&e.remove()}updateToggleButtonIcon(){if(!this.toggleButton||!this.widget)return;const e=this.toggleButton.querySelector(".tms-notification-badge"),t=e?e.outerHTML:"";this.toggleButton.innerHTML=`
      ${this.getBubbleStyleIcon(this.widget.chat_bubble_style)}
      ${t}
    `}connectWebSocket(){if(this.session)try{const e=this.api.getWebSocketUrl(this.session.token,this.session.widget_id);this.websocket=new WebSocket(e),this.websocket.onopen=()=>{console.log("WebSocket connected"),this.reconnectAttempts=0,this.isConnected=!0,this.updateStatus("Connected")},this.websocket.onmessage=t=>{try{const i=JSON.parse(t.data);this.handleWebSocketMessage(i)}catch(i){console.error("Failed to parse WebSocket message:",i)}},this.websocket.onclose=t=>{if(console.log("WebSocket disconnected:",t.code,t.reason),this.isConnected=!1,this.updateStatus(this.isBusinessHoursOpen?"Connecting...":"Away"),this.stopTyping(),this.reconnectAttempts<this.maxReconnectAttempts){this.reconnectAttempts++;const i=Math.min(this.reconnectDelay*Math.pow(2,this.reconnectAttempts-1),3e4);console.log(`Reconnecting in ${i}ms (attempt ${this.reconnectAttempts})`),setTimeout(()=>this.connectWebSocket(),i)}else this.updateStatus("Connection failed")},this.websocket.onerror=t=>{console.error("WebSocket error:",t),this.isConnected=!1,this.updateStatus("Connection error")}}catch(e){console.error("Failed to connect WebSocket:",e),this.isConnected=!1}}handleWebSocketMessage(e){var t;switch(e.type){case"chat_message":this.messages.find(n=>n.id===e.data.id)||(this.addMessage(e.data),this.emitter.emit("message:received",e.data),!this.isOpen&&e.data.author_type==="agent"&&(this.unreadCount++,this.updateNotificationBadge(),(t=this.widget)!=null&&t.sound_enabled&&q("notification",!0)),this.isOpen&&e.data.author_type==="agent"&&this.sendReadReceipt(e.data.id));break;case"agent_joined":this.updateStatus(`${e.data.agent_name} joined`),this.emitter.emit("agent:joined",e.data);const o={id:"agent-joined-"+Date.now(),content:`${e.data.agent_name} has joined the conversation`,author_type:"system",author_name:"System",created_at:new Date().toISOString(),message_type:"text",is_private:!1};this.addMessage(o);break;case"typing_start":e.data.author_type==="agent"&&(this.showTypingIndicator(e.data.author_name),this.emitter.emit("agent:typing",e.data));break;case"typing_stop":this.hideTypingIndicator();break;case"message_read":this.handleMessageRead(e.data.message_id);break;case"session_update":e.data.status==="ended"&&this.handleSessionEnd();break;case"error":this.emitter.emit("error",e.data.error),this.showError(e.data.error);break;default:console.warn("Unknown WebSocket message type:",e.type)}}handleSessionEnd(){this.updateStatus("Session ended");const e={id:"session-ended-"+Date.now(),content:"The conversation has ended. Feel free to start a new chat if you need further assistance.",author_type:"system",author_name:"System",created_at:new Date().toISOString(),message_type:"text",is_private:!1};this.addMessage(e),this.session=null,this.storage.clearSession(),this.websocket&&(this.websocket.close(),this.websocket=null)}addMessage(e){this.messages.push(e),this.displayMessage(e),this.storage.addMessage(e),this.storage.updateSessionActivity()}async sendMessage(){const e=document.getElementById("tms-chat-input");if(!e||!this.session)return;const t=(e.textContent||"").trim();if(t){e.innerHTML="",e.style.height="auto",this.stopTyping();try{const i={id:"temp-"+Date.now(),content:t,author_type:"visitor",author_name:"You",created_at:new Date().toISOString(),message_type:"text",is_private:!1};if(this.addMessage(i),this.isConnected&&this.websocket){const o={type:"chat_message",client_session_id:this.session.id,data:{content:t,message_type:"text",author_type:"visitor",author_name:"You"}};this.websocket.send(JSON.stringify(o)),this.emitter.emit("message:sent",i)}else console.error("WebSocket not connected, not able to send the chat message")}catch(i){console.error("Failed to send message:",i),this.emitter.emit("error","Failed to send message"),this.showError("Failed to send message. Please try again.");const o=this.messages.findIndex(n=>n.id.startsWith("temp-"));o!==-1&&(this.messages.splice(o,1),this.refreshMessages())}}}sendQuickReaction(e){const t=document.getElementById("tms-chat-input");t&&(t.textContent=e,this.sendMessage())}handleTyping(){!this.isConnected||!this.websocket||!this.session||(this.typingTimeout&&(clearTimeout(this.typingTimeout),this.typingTimeout=null),this.isTyping||(this.isTyping=!0,this.sendTypingIndicator(!0)),this.typingTimeout=window.setTimeout(()=>{this.stopTyping()},2e3))}sendTypingIndicator(e){if(!(!this.isConnected||!this.websocket||!this.session))try{const t={type:e?"typing_start":"typing_stop",client_session_id:this.session.id,data:{author_type:"visitor",author_name:"You"}};this.websocket.send(JSON.stringify(t))}catch(t){console.error("Failed to send typing indicator:",t)}}stopTyping(){this.isTyping&&(this.isTyping=!1,this.sendTypingIndicator(!1)),this.typingTimeout&&(clearTimeout(this.typingTimeout),this.typingTimeout=null)}showTypingIndicator(e){const t=document.getElementById("tms-chat-typing"),i=document.getElementById("tms-typing-text");t&&i&&(i.textContent=`${e} is typing...`,t.style.display="flex")}hideTypingIndicator(){const e=document.getElementById("tms-chat-typing");e&&(e.style.display="none")}updateStatus(e){var i;const t=(i=this.container)==null?void 0:i.querySelector(".tms-agent-status");if(t){const o=t.querySelector(".tms-status-indicator");t.innerHTML="",o&&t.appendChild(o),t.appendChild(document.createTextNode(e))}}sendReadReceipt(e){if(!(!this.isConnected||!this.websocket||!this.session))try{const t={type:"message_read",client_session_id:this.session.id,data:{message_id:e,read_by:"visitor"}};this.websocket.send(JSON.stringify(t))}catch(t){console.error("Failed to send read receipt:",t)}}handleMessageRead(e){const t=this.messages.find(i=>i.id===e);t&&t.author_type==="visitor"&&console.log(`Message ${e} was read by agent`)}refreshMessages(){const e=document.getElementById("tms-chat-messages");e&&(e.innerHTML="",this.messages.forEach(t=>this.displayMessage(t)))}escapeHtml(e){const t={"&":"&amp;","<":"&lt;",">":"&gt;",'"':"&quot;","'":"&#039;"};return e.replace(/[&<>"']/g,i=>t[i])}on(e,t){this.emitter.on(e,t)}off(e,t){this.emitter.off(e,t)}destroy(){this.typingTimeout&&(clearTimeout(this.typingTimeout),this.typingTimeout=null),this.stopTyping(),this.websocket&&this.websocket.close(),this.container&&this.container.remove(),this.toggleButton&&this.toggleButton.remove();const e=document.getElementById("tms-widget-styles");e&&e.remove(),this.options.enableSessionPersistence===!1&&this.storage.cleanup()}openWidget(){this.open()}closeWidget(){this.close()}toggleWidget(){this.toggle()}updateWidgetConfig(e){this.widget&&(Object.assign(this.widget,e),xe(this.widget),this.updateToggleButtonIcon())}}function Se(){window.TMSChatConfig&&new Y(window.TMSChatConfig)}document.readyState==="loading"?document.addEventListener("DOMContentLoaded",Se):Se(),window.TMSChatWidget=Y,A.TMSChatWidget=Y,Object.defineProperty(A,Symbol.toStringTag,{value:"Module"})});
