import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { apiRequest, getStoredToken, setStoredToken, setUnauthorizedHandler } from './request';
beforeEach(()=>{localStorage.clear();setStoredToken('token');});
afterEach(()=>vi.unstubAllGlobals());
const respond=(body:unknown,status=200)=>vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response(JSON.stringify(body),{status})));
describe('GVA response adapter',()=>{
 it.each([{data:null},{data:false},{data:0},{data:{}},{data:[]}])('returns data exactly: %j',async ({data})=>{respond({code:0,data,msg:'ok'});expect(await apiRequest('/system/audit-logs')).toEqual(data);});
 it.each([{items:[]},{code:0,msg:'ok'},{code:'0',data:{},msg:'ok'},{code:0,data:{},msg:4},null,[]])('rejects malformed or legacy success: %j',async body=>{respond(body);await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({kind:'protocol',status:200});expect(getStoredToken()).toBe('token');});
 it('rejects HTTP 2xx business failure and retains code',async()=>{respond({code:7,data:{},msg:'not allowed'});await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({kind:'business',status:200,code:7,message:'not allowed'});expect(getStoredToken()).toBe('token');});
 it('does not mistake non-2xx code 0 for success',async()=>{respond({code:0,data:{},msg:'server failed'},500);await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({kind:'http',status:500,code:0,message:'server failed'});});
 it.each([{msg:'new error',code:7,data:{}},{message:'old error'}])('reads both HTTP error formats only in the error branch',async body=>{respond(body,403);await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({status:403,message:'msg' in body?'new error':'old error'});expect(getStoredToken()).toBe('token');});
 it('rejects malformed success JSON without clearing the session',async()=>{vi.stubGlobal('fetch',vi.fn().mockResolvedValue(new Response('{broken')));await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({kind:'protocol',status:200});expect(getStoredToken()).toBe('token');});
 it('expires only authenticated current-token HTTP 401',async()=>{const expired=vi.fn();setUnauthorizedHandler(expired);respond({code:7,data:null,msg:'expired'},401);await expect(apiRequest('/system/audit-logs')).rejects.toMatchObject({status:401,code:7});expect(getStoredToken()).toBeNull();expect(expired).toHaveBeenCalledTimes(1);});
 it('does not clear a session on a public login 401',async()=>{const expired=vi.fn();setUnauthorizedHandler(expired);respond({code:7,data:null,msg:'bad login'},401);await expect(apiRequest('/system/auth/login',{auth:'none'})).rejects.toMatchObject({status:401,code:7});expect(getStoredToken()).toBe('token');expect(expired).not.toHaveBeenCalled();});
});
